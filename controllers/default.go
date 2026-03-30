package controllers

import (
	"GanzamApi/config"
	beego "github.com/astaxie/beego"
)

const CurrentVersion = "v1"

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	c.Data["Website"] = "beego.vip"
	c.Data["Email"] = "astaxie@gmail.com"
	c.TplName = "index.tpl"
}

type VersionController struct {
	beego.Controller
}

func (c *VersionController) Get() {
	c.Data["json"] = map[string]string{
		"version":    CurrentVersion,
		"environment": config.GetAppEnv(),
		"target_url":  config.GetTargetURL(),
	}
	c.ServeJSON()
}
