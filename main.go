package main

import (
	conf "GanzamApi/conf"
	_ "GanzamApi/routers"
	"fmt"
	"os"
	"strconv"
	"strings"

	beego "github.com/astaxie/beego"
	beecontext "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/plugins/cors"
)

func main() {
	beego.SetLevel(beego.LevelDebug)
	beego.BeeLogger.SetLogger("console")

	conf.Env = conf.GetAppEnv()

	configureHTTPListen()

	beego.Info(fmt.Sprintf("DB target: %s", conf.GetDBTargetSummary()))

	beego.InsertFilter("*", beego.FinishRouter, func(ctx *beecontext.Context) {
		beego.Info("%s %s -> %d", ctx.Input.Method(), ctx.Input.URL(), ctx.ResponseWriter.Status)
	}, true)

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     getCORSAllowOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	beego.Run()
}

func configureHTTPListen() {
	if value := strings.TrimSpace(os.Getenv("HTTP_ADDR")); value != "" {
		beego.BConfig.Listen.HTTPAddr = value
	} else {
		beego.BConfig.Listen.HTTPAddr = "0.0.0.0"
	}

	port := strings.TrimSpace(os.Getenv("HTTP_PORT"))
	if port == "" {
		port = strings.TrimSpace(os.Getenv("PORT"))
	}
	if port == "" {
		beego.BConfig.Listen.HTTPPort = 8081
		return
	}

	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort <= 0 {
		beego.Warn(fmt.Sprintf("invalid HTTP port %q, using 8080", port))
		beego.BConfig.Listen.HTTPPort = 8081
		return
	}

	beego.BConfig.Listen.HTTPPort = parsedPort
}

func getCORSAllowOrigins() []string {
	if value := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS")); value != "" {
		parts := strings.Split(value, ",")
		origins := make([]string, 0, len(parts))
		for _, part := range parts {
			if origin := strings.TrimSpace(part); origin != "" {
				origins = append(origins, origin)
			}
		}
		if len(origins) > 0 {
			return origins
		}
	}

	return []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://172.30.30.30:5173",
		"http://35.74.65.223",
		"http://35.74.65.223:5173",
	}
}
