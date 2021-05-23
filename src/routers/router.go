package routers

import (
	"beego-goodsm/controllers"
	_ "beego-goodsm/models"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/api", &controllers.MainController{})
	beego.Router("/", &controllers.IndexController{})
	beego.Router("/add", &controllers.AddGoodController{})
}
