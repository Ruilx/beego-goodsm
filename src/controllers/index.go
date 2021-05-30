package controllers

import (
	"beego-goodsm/common"
	"beego-goodsm/models"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"strings"
)

type IndexController struct {
	beego.Controller
	log *logs.BeeLogger
}

func (c *IndexController) Prepare() {
	c.log = logs.GetBeeLogger()
}

func (c *IndexController) Get() {
	name := c.Ctx.Input.Query("name")
	if name != ""{
		name = strings.Trim(name, " ")
	}

	ml, err := models.GetGoods(name, models.ORDERBY_DESC)
	if err != nil{
		c.log.Error("[DB] database error: " + err.Error())
		c.Data["err"] = "系统错误, 请检查系统日志文件输出"
		return
	}
	var goods []map[string]interface{}
	for _, g := range ml {
		good, err := common.Struct2Map(g, true)
		if err != nil{
			c.log.Warn("[Struct2Map] transform failed. error: " + err.Error() + ", data ignored.")
			continue
		}
		goods = append(goods, good)
	}
	c.Data["goods"] = goods

	c.TplName = "index.html"
}
