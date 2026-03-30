package main

import (
	"GanzamApi/config"
	_ "GanzamApi/routers"
	beego "github.com/astaxie/beego"
)

func main() {
	beego.BConfig.RunMode = config.GetAppEnv()
	beego.Run()
}
