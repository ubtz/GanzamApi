package controllers

import (
	"io"
	"net/http"

	beego "github.com/astaxie/beego"
)

type UploadImage struct {
	beego.Controller
}

func (c *UploadImage) Post() {
	if !requireImageUploadService(&c.Controller) {
		return
	}

	if err := c.Ctx.Request.ParseMultipartForm(10 << 20); err != nil {
		abortJSON(&c.Controller, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := c.GetFile("image")
	if err != nil {
		abortJSON(&c.Controller, http.StatusBadRequest, "image file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		abortJSON(&c.Controller, http.StatusInternalServerError, "failed to read image")
		return
	}

	result, err := uploadService.UploadImage(
		c.Ctx.Request.Context(),
		header.Filename,
		header.Header.Get("Content-Type"),
		data,
	)
	if err != nil {
		beego.Error("upload image failed:", err)
		abortJSON(&c.Controller, http.StatusInternalServerError, "upload failed")
		return
	}

	c.Data["json"] = result
	c.ServeJSON()
}
