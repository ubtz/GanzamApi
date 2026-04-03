package controllers

import (
	conf "GanzamApi/conf"
	beego "github.com/astaxie/beego"
)

const CurrentVersion = "v1"

type GetVersion struct {
	beego.Controller
}

func (c *GetVersion) GetVersion() {
	c.Data["json"] = map[string]string{
		"version":     CurrentVersion,
		"environment": conf.GetAppEnv(),
		"target_url":  conf.GetTargetURL(),
	}
	c.ServeJSON()
}
