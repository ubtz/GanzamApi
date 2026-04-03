package main

import (
	conf "GanzamApi/conf"
	_ "GanzamApi/routers"

	beego "github.com/astaxie/beego"
	beecontext "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/plugins/cors"
)

func main() {
	beego.SetLevel(beego.LevelDebug)
	beego.BeeLogger.SetLogger("console")
	conf.Env = conf.GetAppEnv()
	beego.Info("DB target: %s", conf.GetDBTargetSummary())

	// Log every request after the handler runs so the final status code is available.
	beego.InsertFilter("*", beego.FinishRouter, func(ctx *beecontext.Context) {
		beego.Info("%s %s -> %d", ctx.Input.Method(), ctx.Input.URL(), ctx.ResponseWriter.Status)
	}, true)

	// CORS
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"}, // or "*"
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	beego.Run()
}
