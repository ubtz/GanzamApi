package controllers

import (
	"context"
	"net/http"

	"GanzamApi/services"
	beego "github.com/astaxie/beego"
)

type imageUploadService interface {
	UploadImage(ctx context.Context, fileName string, contentType string, data []byte) (*services.UploadImageResult, error)
}

var uploadService imageUploadService

func init() {
	service, err := services.NewS3ImageService(context.Background())
	if err != nil {
		beego.Warn("failed to initialize upload service:", err)
		return
	}

	uploadService = service
}

func SetImageUploadService(service imageUploadService) {
	uploadService = service
}

func requireImageUploadService(c *beego.Controller) bool {
	if uploadService != nil {
		return true
	}

	abortJSON(c, http.StatusInternalServerError, "upload service is not configured")
	return false
}
