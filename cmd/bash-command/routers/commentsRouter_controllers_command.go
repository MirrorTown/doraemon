package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {
	const CommandController = "github.com/Qihoo360/doraemon/cmd/bash-command/controllers/command:CommandController"
	beego.GlobalControllerRouter[CommandController] = append(
		beego.GlobalControllerRouter[CommandController],
		beego.ControllerComments{
			Method:           "Exec",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil,
		})
}
