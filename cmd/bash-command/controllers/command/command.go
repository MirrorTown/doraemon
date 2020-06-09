package command

import (
	"encoding/json"
	"github.com/Qihoo360/doraemon/cmd/alert-gateway/logs"
	"github.com/Qihoo360/doraemon/cmd/bash-command/api"
	"github.com/Qihoo360/doraemon/cmd/bash-command/common/base"
)

type CommandController struct {
	base.APIController
}

type Args struct {
	Name  string `json:"name"`
	Shell string `json:"shell"`
}

func (c *CommandController) URLMapping() {
	c.Mapping("Exec", c.Exec)
}

func (c *CommandController) Prepare() {
	c.APIController.Prepare()

}

// @Title Exec
// @Description get all objects
// @Success 200 {object} int64 success
// @router / [post]
func (c *CommandController) Exec() {
	var body = new(Args)
	err := json.Unmarshal(c.Ctx.Input.RequestBody, body)
	if err != nil {
		logs.Error("解析请求body失败")
		c.AbortBadRequest("请求体格式错误")
	}

	var client = api.Command{Name: body.Name}
	go client.Exec(body.Shell)

	c.Success("ok")
}
