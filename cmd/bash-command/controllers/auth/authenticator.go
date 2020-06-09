package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Qihoo360/doraemon/cmd/alert-gateway/logs"
	rsakey "github.com/Qihoo360/doraemon/cmd/bash-command/apikey"
	"github.com/Qihoo360/doraemon/cmd/bash-command/common/base"
	"github.com/Qihoo360/doraemon/cmd/bash-command/common/sso"
	"github.com/Qihoo360/doraemon/cmd/bash-command/models"
	"github.com/Qihoo360/doraemon/cmd/bash-command/util/hack"
	"github.com/astaxie/beego"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

// Authenticator provides interface to authenticate user credentials.
type Authenticator interface {
	// Authenticate ...
	Authenticate(m models.AuthModel) (*models.User, error)
}

var registry = make(map[string]Authenticator)

// Register add different authenticators to registry map.
func Register(name string, authenticator Authenticator) {
	if _, dup := registry[name]; dup {
		logs.Info("authenticator: %s has been registered", name)
		return
	}
	registry[name] = authenticator
}

// AuthController operations for Auth
type AuthController struct {
	beego.Controller
}

// URLMapping ...
func (c *AuthController) URLMapping() {
	c.Mapping("Login", c.Login)
	c.Mapping("Logout", c.Logout)
	c.Mapping("CurrentUser", c.CurrentUser)
}

type LoginResult struct {
	Token string `json:"token"`
}

// type is login type
// name when login type is oauth2 used for oauth2 type
// @router /login/:type/?:name [get,post]
func (c *AuthController) Login() {
	username := c.Input().Get("username")
	password := c.Input().Get("password")
	authType := c.Ctx.Input.Param(":type")
	oauth2Name := c.Ctx.Input.Param(":name")
	next := c.Ctx.Input.Query("next")
	if authType == "" || username == "admin" {
		authType = models.AuthTypeDB
	}
	logs.Info("auth type is", authType)
	authenticator, ok := registry[authType]
	if !ok {
		logs.Warning("auth type (%s) is not supported . ", authType)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Ctx.Output.Body(hack.Slice(fmt.Sprintf("auth type (%s) is not supported.", authType)))
		return
	}
	authModel := models.AuthModel{
		Username: username,
		Password: password,
	}

	var userinfo *sso.BasicUserInfo
	var err error
	if authType == models.AuthTypeSso {
		ssoer, ok := sso.SsoInfos[oauth2Name]
		if !ok {
			logs.Warning("sso type (%s) is not supported . ", oauth2Name)
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Ctx.Output.Body(hack.Slice("sso type is not supported."))
			return
		}
		userinfo, err = c.getUserInfo(ssoer)
		if err != nil || userinfo == nil {
			logs.Error("获取用户信息失败,", err)
			return
		}
	}

	//if authType == models.AuthTypeOAuth2 {
	//	oauther, ok := selfoauth.OAutherMap[oauth2Name]
	//	if !ok {
	//		logs.Warning("oauth2 type (%s) is not supported . ", oauth2Name)
	//		c.Ctx.Output.SetStatus(http.StatusBadRequest)
	//		c.Ctx.Output.Body(hack.Slice("oauth2 type is not supported."))
	//		return
	//	}
	//	code := c.Input().Get("code")
	//	if code == "" {
	//		c.Ctx.Redirect(http.StatusFound, oauther.AuthCodeURL(next, oauth2.AccessTypeOnline))
	//		return
	//	}
	//	authModel.OAuth2Code = code
	//	authModel.OAuth2Name = oauth2Name
	//	state := c.Ctx.Input.Query("state")
	//	if state != "" {
	//		next = state
	//	}
	//
	//}

	var user = new(models.User)
	if authType == models.AuthTypeSso {
		//var ssoAuth sso.SSoAuth
		user, err = sso.Authenticate(userinfo)
	} else {
		user, err = authenticator.Authenticate(authModel)
	}
	if err != nil || user == nil {
		logs.Warning("try to login in with user (%s) error %v. ", authModel.Username, err)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Ctx.Output.Body(hack.Slice(fmt.Sprintf("Login failed. %v", err)))
		return
	}

	now := time.Now()
	user.LastIp = RemoteIp(c.Ctx.Request)
	user.LastLogin = &now
	user, err = models.UserModel.EnsureUser(user)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Ctx.Output.Body(hack.Slice(err.Error()))
		return
	}

	// default token exp time is 3600s.
	expSecond := beego.AppConfig.DefaultInt64("TokenLifeTime", 86400)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		// 签发者
		"iss": "wayne",
		// 签发时间
		"iat": now.Unix(),
		"exp": now.Add(time.Duration(expSecond) * time.Second).Unix(),
		"aud": user.Name,
	})

	apiToken, err := token.SignedString(rsakey.RsaPrivateKey)
	if err != nil {
		logs.Error("create token form rsa private key  error.", rsakey.RsaPrivateKey, err.Error())
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Ctx.Output.Body(hack.Slice(err.Error()))
		return
	}

	if next != "" {
		// if oauth type is oauth, set token for client.
		if authType == models.AuthTypeOAuth2 || authType == models.AuthTypeSso {
			next = next + "&sid=" + apiToken
		}
		c.Redirect(next, http.StatusFound)
		return
	}

	loginResult := LoginResult{
		Token: apiToken,
	}
	c.Data["json"] = base.Result{Data: loginResult}
	c.ServeJSON()
}

func RemoteIp(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get("XRealIP"); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get("XForwardedFor"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

func (c *AuthController) getUserInfo(s *sso.SsoInfo) (*sso.BasicUserInfo, error) {
	var userInfo = new(sso.BasicUserInfo)
	//获取用户信息
	token := c.Ctx.Input.Cookie("_security_token_inc")
	if token == "" {
		loginurl := s.RedirectUrl + base64.URLEncoding.EncodeToString([]byte(s.BackUrl))
		c.Ctx.Redirect(http.StatusFound, loginurl)
		return nil, nil
	} else {
		client := &http.Client{}
		var req *http.Request
		req, _ = http.NewRequest(http.MethodGet, s.GetAuth, nil)
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(req.URL, []*http.Cookie{
			&http.Cookie{Name: "_security_token_inc", Value: c.Ctx.Input.Cookie("_security_token_inc")},
			&http.Cookie{Name: "isDingDing", Value: c.Ctx.Input.Cookie("isDingDing")},
			&http.Cookie{Name: "tracknick", Value: c.Ctx.Input.Cookie("tracknick")},
		})
		client.Jar = jar
		resp, err := client.Do(req)
		if err != nil {
			logs.Error("获取用户信息失败,", err)
		}
		b, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		var mapResult map[string]interface{}
		err = json.Unmarshal(b, &mapResult)
		if err != nil {
			logs.Error("JsonToMapDemo err: ", err)
		}
		userInfo.Name = mapResult["data"].(map[string]interface{})["userName"].(string)
		userInfo.Email = mapResult["data"].(map[string]interface{})["email"].(string)
		userInfo.Display = mapResult["data"].(map[string]interface{})["displayName"].(string)
	}

	return userInfo, nil
}

// @router /logout [get]
func (c *AuthController) Logout() {

}

// @router /currentuser [get]
func (c *AuthController) CurrentUser() {
	c.Controller.Prepare()
	authString := c.Ctx.Input.Header("Authorization")

	kv := strings.Split(authString, " ")
	if len(kv) != 2 || kv[0] != "Bearer" {
		logs.Info("AuthString invalid:", authString)
		c.CustomAbort(http.StatusUnauthorized, "Token Invalid ! ")
	}
	tokenString := kv[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return rsakey.RsaPublicKey, nil
	})
	errResult := base.ErrorResult{}
	switch err.(type) {
	case nil: // no error
		if !token.Valid { // but may still be invalid
			errResult.Code = http.StatusUnauthorized
			errResult.Msg = "Token Invalid ! "
		}

	case *jwt.ValidationError: // something was wrong during the validation
		errResult.Code = http.StatusUnauthorized
		errResult.Msg = err.Error()

	default: // something else went wrong
		errResult.Code = http.StatusInternalServerError
		errResult.Msg = err.Error()
	}

	if err != nil {
		c.CustomAbort(errResult.Code, errResult.Msg)
	}

	claim := token.Claims.(jwt.MapClaims)
	aud := claim["aud"].(string)
	user, err := models.UserModel.GetUserDetail(aud)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, err.Error())
	}

	c.Data["json"] = base.Result{Data: user}
	c.ServeJSON()
}
