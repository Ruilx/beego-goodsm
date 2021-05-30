package controllers

import (
	"beego-goodsm/models"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"strconv"
)

type AddGoodController struct {
	beego.Controller
	log *logs.BeeLogger
}

func (c *AddGoodController) Prepare() {
	c.log = logs.GetBeeLogger()
}

func (c *AddGoodController) Get() {
	var err error
	var idi int
	id := c.Ctx.Input.Query("id")
	if id != ""{
		if idi, err = strconv.Atoi(id); err != nil{
			c.log.Error("input argument string is not an integer: " + err.Error() + ", id = " + id[:100])
			idi = -1
		}
	}
	c.TplName = "addgood.html"
	if idi > 0{
		good, err := models.GetGoodsById(int32(idi))
		if err == orm.ErrNoRows{
			c.log.Error("database not found good id = " + id + ".")
			return
		}
		if err != nil && err != orm.ErrNoRows{
			c.log.Error("database result error: " + err.Error())
			return
		}
		if good != nil {
			c.Data["id"] = good.Id
			c.Data["name"] = good.Name
			c.Data["desc"] = good.Desc
			c.Data["price"] = good.Price
			c.Data["quantity"] = good.Quantity
			c.Data["image"] = good.Image
		}
	}
}