package routers

import (
	"github.com/Qihoo360/doraemon/cmd/bash-command/controllers/auth"
	"github.com/Qihoo360/doraemon/cmd/bash-command/controllers/command"
	"github.com/Qihoo360/doraemon/cmd/bash-command/health"
	"github.com/Qihoo360/doraemon/cmd/bash-command/util/hack"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/plugins/cors"
	"net/http"
	"path"
)

func init() {
	// Beego注解路由代码生成规则和程序运行路径相关，需要改写一下避免产生不一致的文件名
	if beego.BConfig.RunMode == "dev" && path.Base(beego.AppPath) == "_build" {
		beego.AppPath = path.Join(path.Dir(beego.AppPath), "cmd/bash-command")
	}

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	beego.Get("/healthz", func(ctx *context.Context) {
		dc := health.EnsureDatabase{}
		err := dc.Health()
		if err != nil {
			ctx.Output.SetStatus(http.StatusInternalServerError)
			ctx.Output.Body(hack.Slice(err.Error()))
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.Body(hack.Slice("ok"))
	})

	beego.Include(&auth.AuthController{})

	nsWithCommand := beego.NewNamespace("/api/v1",
		beego.NSNamespace("/command",
			beego.NSInclude(&command.CommandController{}),
		),
	)

	beego.AddNamespace(nsWithCommand)
}
