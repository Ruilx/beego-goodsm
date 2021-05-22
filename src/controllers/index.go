package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
)

type IndexController struct {
	beego.Controller
}

func (c *IndexController) Prepare() {

}

func (c *IndexController) Get() {
	name := c.Ctx.Input.Query("name")

	c.TplName = "index.html"
}
