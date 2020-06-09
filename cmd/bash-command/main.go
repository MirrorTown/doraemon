package main

import (
	"github.com/Qihoo360/doraemon/cmd/bash-command/initial"
	"github.com/Qihoo360/doraemon/cmd/bash-command/models"
	_ "github.com/Qihoo360/doraemon/cmd/bash-command/routers"
	"github.com/astaxie/beego"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	//init rsaKey
	initial.InitRsaKey()

	//init database
	models.InitDb()

	//init and start websocket
	ws := initial.NewWsServer()
	go ws.Start()

	beego.Run()
}
