package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"GanzamApi/models"
	"GanzamApi/services"
	beego "github.com/astaxie/beego"
)

type Register struct {
	beego.Controller
}

func (c *Register) PostRegister() {
	if !requireAuthService(&c.Controller) {
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&req); err != nil {
		abortJSON(&c.Controller, http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := authService.Register(c.Ctx.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserExists):
			abortJSON(&c.Controller, http.StatusConflict, err.Error())
		case err.Error() == "phone and password are required":
			abortJSON(&c.Controller, http.StatusBadRequest, err.Error())
		default:
			beego.Error("register failed:", err)
			abortJSON(&c.Controller, http.StatusInternalServerError, "register failed")
		}
		return
	}

	c.Data["json"] = map[string]interface{}{
		"token": token,
		"user":  user,
	}
	c.ServeJSON()
}
