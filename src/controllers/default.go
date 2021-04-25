package controllers

import (
	"beego-goodsm/models"
	"fmt"
	beego "github.com/beego/beego/v2/server/web"
	"reflect"
	"strconv"
)

type ResJson struct {
	Code int32 `json:"code"`
	Msg string `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

type MainController struct {
	beego.Controller
	res ResJson
}

func (c *MainController) Prepare(){
	if !c.IsAjax() {
		//c.Ctx.Abort(400, "No current url found.")
		c.Data["ajaxreq"] = "Not in ajax"
		c.Ctx.Output.Header("ajaxreq", "notforce")
	}
	c.res.Data = make(map[string]interface{})
	c.res.Msg = "Success."
}

//func (c *MainController) Get() {
//	c.Ctx.WriteString("Welcome to use api!")
//}

func (c *MainController) DoServeJSON(){
	c.Data["json"] = c.res
	if err := c.ServeJSON(); err != nil {
		fmt.Println("Cannot serve json string")
	}
}

func (c *MainController) AjaxSetResult(code int32, msg string){
	//c.Ctx.Output.Status = statusCode
	c.res.Code = code
	c.res.Msg = msg
}

func (c *MainController) Get() {

	defer c.DoServeJSON()
	op := c.Ctx.Input.Query("op")

	switch op{
	case "add": // Add Good
		c.AddGood()
	case "upd": // Update Good
	case "del": // Delete Good
	case "sel": // Sell Good
	case "sto": // Storaging Goods
	default:
		c.res.Msg = "Not a valid operation"
	}

	fmt.Println(c.res)
}

func (c *MainController) AddGood(){
	name := c.Ctx.Input.Query("name")
	desc := c.Ctx.Input.Query("desc")
	price := c.Ctx.Input.Query("price")
	quantity := c.Ctx.Input.Query("quantity")

	fmt.Println(reflect.TypeOf(name), "=", name, ", ",
		reflect.TypeOf(desc), "=", desc, ", ",
		reflect.TypeOf(price), "=", price, ", ",
		reflect.TypeOf(quantity), "=", quantity)

	if name == "" {
		c.AjaxSetResult(400, "Expect an argument 'name', but not found.")
		return
	}
	var floatPrice float64
	var err error
	var intQuantity int64
	var id int64
	if price == "" {
		floatPrice = 0
	}
	if floatPrice, err = strconv.ParseFloat(price, 64); err != nil {
		c.AjaxSetResult(400, "Expect an float type 'price', but parse failed: " + err.Error())
		return
	}
	if quantity == "" {
		intQuantity = 0
	}
	if intQuantity, err = strconv.ParseInt(quantity, 10, 64); err != nil {
		c.AjaxSetResult(400, "Expect an int type 'quantity', but parse failed: " + err.Error())
	}
	good := models.Good{Name: name, Desc: desc, Price: floatPrice, Quantity: intQuantity}
	if id, err = models.AddGoods(&good); err != nil {
		c.AjaxSetResult(501, "Cannot insert a good to database, system error: " + err.Error())
	}
	c.res.Data["id"] = id
	c.AjaxSetResult(200, "Success")
}

