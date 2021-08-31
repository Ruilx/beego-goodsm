package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"strings"
)

type HistoryController struct{
	beego.Controller
	log *logs.BeeLogger
}

func (c *HistoryController) Prepare(){
	c.log = logs.GetBeeLogger()
}

func (c *HistoryController) Get(){
	//var err error
	name := c.Ctx.Input.Query("name")
	if name != ""{
		name = strings.Replace(name, "'", "\\'", -1)
	}

	c.TplName = "history.html"
}