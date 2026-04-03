package controllers

import (
	"net/http"

	"GanzamApi/repositories"
	"GanzamApi/services"
	beego "github.com/astaxie/beego"
)

var authService *services.AuthService

func init() {
	store, err := repositories.NewMSSQLUserStore()
	if err == nil {
		authService = services.NewAuthService(store)
		return
	}

	beego.Error("failed to initialize auth service:", err)
}

func SetAuthService(service *services.AuthService) {
	authService = service
}

func abortJSON(c *beego.Controller, status int, message string) {
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = map[string]string{
		"error": message,
	}
	c.ServeJSON()
}

func requireAuthService(c *beego.Controller) bool {
	if authService != nil {
		return true
	}

	abortJSON(c, http.StatusInternalServerError, "auth service is not configured")
	return false
}
