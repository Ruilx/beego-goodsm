package controllers

import (
	"beego-goodsm/models"
	"fmt"
	"reflect"
	"strconv"

	beego "github.com/beego/beego/v2/server/web"
)

type ResJson struct {
	Code int32                  `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

type MainController struct {
	beego.Controller
	res ResJson
}

func (c *MainController) Prepare() {
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

func (c *MainController) DoServeJSON() {
	c.Data["json"] = c.res
	if err := c.ServeJSON(); err != nil {
		fmt.Println("Cannot serve json string")
	}
}

func (c *MainController) AjaxSetResult(code int32, msg string) {
	//c.Ctx.Output.Status = statusCode
	c.res.Code = code
	c.res.Msg = msg
}

func (c *MainController) Post() {

	defer c.DoServeJSON()
	op := c.Ctx.Input.Query("op")

	switch op {
	case "add": // Add Good
		c.AddGood()
		break
	case "upd": // Update Good
	case "del": // Delete Good
	case "sel": // Sell Good
	case "get": // Get Goods
		c.GetGoods()
		break
	default:
		c.res.Msg = "Not a valid operation"
	}

	fmt.Println(c.res)
}

func (c *MainController) GetGoods() {
	name := c.Ctx.Input.Query("name")

	fmt.Println("Search: " + name)

	var find_fields map[string]string
	if name != "" {
		find_fields["name"] = name
	}

	var goods []interface{}
	var find []string
	var sort []int
	var err error
	if goods, err = models.GetGoods(find_fields, find, find, sort, 0, 0); err != nil {
		c.AjaxSetResult(500, "Expect to read database but failed: "+err.Error())
		return
	}
	fmt.Println(goods)

}

func (c *MainController) AddGood() {
	name := c.Ctx.Input.Query("name")
	desc := c.Ctx.Input.Query("desc")
	price := c.Ctx.Input.Query("price")
	quantity := c.Ctx.Input.Query("quantity")

	var err error

	fmt.Println(c.Ctx.Input.Query("image"))
	return
	file, handle, err := c.GetFile("image")

	fmt.Println(reflect.TypeOf(name), "=", name, ", ",
		reflect.TypeOf(desc), "=", desc, ", ",
		reflect.TypeOf(price), "=", price, ", ",
		reflect.TypeOf(quantity), "=", quantity)

	if name == "" {
		c.AjaxSetResult(400, "Expect an argument 'name', but not found.")
		return
	}
	if err != nil {
		c.AjaxSetResult(500, "Image not a vaild err: "+err.Error())
		return
	}
	defer file.Close()
	imagePath := "static/upload/" + handle.Filename
	c.SaveToFile("image", imagePath)

	var floatPrice float64

	var intQuantity int64
	var id int64
	if price == "" {
		floatPrice = 0
	}
	if floatPrice, err = strconv.ParseFloat(price, 64); err != nil {
		c.AjaxSetResult(400, "Expect an float type 'price', but parse failed: "+err.Error())
		return
	}
	if quantity == "" {
		intQuantity = 0
	}
	if intQuantity, err = strconv.ParseInt(quantity, 10, 64); err != nil {
		c.AjaxSetResult(400, "Expect an int type 'quantity', but parse failed: "+err.Error())
		return
	}
	good := models.Good{Name: name, Desc: desc, Price: floatPrice, Quantity: intQuantity}
	if id, err = models.AddGoods(&good); err != nil {
		c.AjaxSetResult(501, "Cannot insert a good to database, system error: "+err.Error())
		return
	}
	fmt.Println("con: ", id, err == nil)
	c.res.Data["id"] = id
	c.AjaxSetResult(200, "Success")
}
