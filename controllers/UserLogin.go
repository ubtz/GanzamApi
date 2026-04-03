package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"GanzamApi/models"
	"GanzamApi/services"
	beego "github.com/astaxie/beego"
)

type UserLogin struct {
	beego.Controller
}

func (c *UserLogin) PostLogin() {
	if !requireAuthService(&c.Controller) {
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&req); err != nil {
		abortJSON(&c.Controller, http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := authService.Login(c.Ctx.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			abortJSON(&c.Controller, http.StatusUnauthorized, err.Error())
		case err.Error() == "phone and password are required":
			abortJSON(&c.Controller, http.StatusBadRequest, err.Error())
		default:
			beego.Error("login failed:", err)
			abortJSON(&c.Controller, http.StatusInternalServerError, "login failed")
		}
		return
	}

	c.Data["json"] = map[string]interface{}{
		"token": token,
		"user":  user,
	}
	c.ServeJSON()
}
