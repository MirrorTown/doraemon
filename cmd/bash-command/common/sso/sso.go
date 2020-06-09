package sso

import (
	"github.com/Qihoo360/doraemon/cmd/bash-command/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"strconv"
)

func init() {
	NewSsoService()
}

var (
	SsoInfos = make(map[string]*SsoInfo)
)

const (
	SsoTypeDefault = "sso"
)

type BasicUserInfo struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Display string `json:"display"`
}

type SsoInfo struct {
	BackUrl     string
	CookieName  string
	RedirectUrl string // get user info
	GetAuth     string
	Enabled     bool
}

func Authenticate(m *BasicUserInfo) (*models.User, error) {
	username := m.Name
	olduser, err := models.UserModel.GetUserByName(username)
	//无用户名记录情况，同步sso信息到db
	if err == orm.ErrNoRows {
		var user = new(models.User)
		user.Name = m.Name
		user.Email = m.Email
		user.Display = m.Display
		user.Comment = "信息来源于SSO"
		_, err := models.UserModel.AddUser(user)
		if err != nil {
			logs.Error("保存用户信息失败, ", err)
		}
		olduser = user
	}

	return olduser, nil
}

func NewSsoService() {
	allSso := []string{SsoTypeDefault}
	err := beego.LoadAppConfig("ini", "conf/app.conf")
	if err != nil {
		err = beego.LoadAppConfig("ini", "/opt/wayne/conf/app.conf")
		if err != nil {
			panic(err)
		}
	}

	for _, name := range allSso {
		section, err := beego.AppConfig.GetSection("auth." + name)
		if err != nil {
			logs.Info("can't enable sso"+name, err)
			continue
		}
		enabled, err := strconv.ParseBool(section["enabled"])
		if err != nil {
			logs.Info("parse enabled oauth error", err)
			continue
		}

		if !enabled {
			continue
		}

		info := &SsoInfo{
			BackUrl:     section["back_url"],
			CookieName:  section["cookie_name"],
			RedirectUrl: section["redirect_url"],
			GetAuth:     section["get_auth"],
			Enabled:     enabled,
		}

		SsoInfos[SsoTypeDefault] = info

	}
}
