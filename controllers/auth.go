package controllers

import (
	"encoding/json"
	"net/http"

	"GanzamApi/models"
	"GanzamApi/repositories"
	"GanzamApi/services"
	beego "github.com/astaxie/beego"
)

var authService *services.AuthService

func init() {
	store, err := repositories.NewMSSQLUserStore()
	if err == nil {
		authService = services.NewAuthService(store)
	}
}

func SetAuthService(service *services.AuthService) {
	authService = service
}

type AuthController struct {
	beego.Controller
}

func (c *AuthController) Register() {
	if authService == nil {
		c.abortJSON(http.StatusInternalServerError, "auth service is not configured")
		return
	}

	var req models.RegisterRequest
	if err := c.parseJSON(&req); err != nil {
		c.abortJSON(http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := authService.Register(c.Ctx.Request.Context(), req)
	if err != nil {
		switch err {
		case services.ErrUserExists:
			c.abortJSON(http.StatusConflict, err.Error())
		default:
			c.abortJSON(http.StatusBadRequest, err.Error())
		}
		return
	}

	c.Data["json"] = map[string]interface{}{
		"token": token,
		"user":  user,
	}
	c.ServeJSON()
}

func (c *AuthController) Login() {
	if authService == nil {
		c.abortJSON(http.StatusInternalServerError, "auth service is not configured")
		return
	}

	var req models.LoginRequest
	if err := c.parseJSON(&req); err != nil {
		c.abortJSON(http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := authService.Login(c.Ctx.Request.Context(), req)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			c.abortJSON(http.StatusUnauthorized, err.Error())
		default:
			c.abortJSON(http.StatusBadRequest, err.Error())
		}
		return
	}

	c.Data["json"] = map[string]interface{}{
		"token": token,
		"user":  user,
	}
	c.ServeJSON()
}

func (c *AuthController) parseJSON(target interface{}) error {
	return json.NewDecoder(c.Ctx.Request.Body).Decode(target)
}

func (c *AuthController) abortJSON(status int, message string) {
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = map[string]string{
		"error": message,
	}
	c.ServeJSON()
}
