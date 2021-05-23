package controllers

import beego "github.com/beego/beego/v2/server/web"

type AddGoodController struct {
	beego.Controller
}

func (c *AddGoodController) Prepare() {

}

func (c *AddGoodController) Get() {

	c.TplName = "addgood.html"
}