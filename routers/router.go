package routers

import (
	"GanzamApi/controllers"
	beego "github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})

	getNs := beego.NewNamespace("/get",
		beego.NSRouter("/version", &controllers.GetVersion{}, "get:GetVersion"),
	)
	beego.AddNamespace(getNs)

	postNs := beego.NewNamespace("/post",
		beego.NSRouter("/register", &controllers.Register{}, "post:PostRegister"),
		beego.NSRouter("/login", &controllers.UserLogin{}, "post:PostLogin"),
	)
	beego.AddNamespace(postNs)

	// Compatibility aliases for existing clients.
	beego.Router("/version", &controllers.GetVersion{}, "get:GetVersion")
	beego.Router("/api/v1/auth/register", &controllers.Register{}, "post:PostRegister")
	beego.Router("/api/v1/auth/login", &controllers.UserLogin{}, "post:PostLogin")
}
