package routers

import (
	"WEB/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	// * is anything such as post、delete、put, it`s the restAPI mode geshi
	beego.Router("/api/sayhello", &controllers.SayController{}, "*:Sayhello")
	beego.Router("/api/gethello", &controllers.SayController{}, "get,post:Gethello")
	beego.Router("/api/deletehello", &controllers.SayController{}, "get:Sayhello;post:Gethello")
	beego.Router("/api/testModel", &controllers.TestModelController{}, "get,post:TestModel")
	beego.Router("/api/testHttpLib", &controllers.TestModelController{}, "get,post:HttpLibTest")
}
