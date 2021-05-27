package controllers

import (
	"beego-goodsm/common"
	"beego-goodsm/models"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"strconv"
)

type ResJson struct {
	Code int32                  `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

type MainController struct {
	beego.Controller
	res ResJson
	log *logs.BeeLogger
}

func (c *MainController) Prepare() {
	if !c.IsAjax() {
		//c.Ctx.Abort(400, "No current url found.")
		c.Data["ajaxreq"] = "Not in ajax"
		c.Ctx.Output.Header("ajaxreq", "notforce")
	}
	c.res.Data = make(map[string]interface{})
	c.res.Msg = "Return code not mentioned."
	c.log = logs.GetBeeLogger()
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
	c.AjaxSetResult(100, "GetGoods Method.")
	// name := c.Ctx.Input.Query("name")

	// fmt.Println("Search: " + name)

	// var find_fields map[string]string
	// if name != "" {
	// 	find_fields["name"] = name
	// }

	// var goods []interface{}
	// var find []string
	// var sort []int
	// var err error
	// if goods, err = models.GetGoods(find_fields, find, find, sort, 0, 0); err != nil {
	// 	c.AjaxSetResult(500, "Expect to read database but failed: "+err.Error())
	// 	return
	// }
	// fmt.Println(goods)

}

func (c *MainController) AddGood() {
	name := c.Ctx.Input.Query("name")
	desc := c.Ctx.Input.Query("desc")
	price := c.Ctx.Input.Query("price")
	quantity := c.Ctx.Input.Query("quantity")
	hasImg := c.Ctx.Input.Query("hasImg")
	imageFile, imageHeader, imageErr := c.GetFile("image")
	imageSaveFilename := ""

	c.log.Info("Entry AddGood operation.")
	c.log.Info("PARAM: name = '" + name + "'")
	c.log.Info("PARAM: desc = '" + desc + "'")
	c.log.Info("PARAM: price = '" + price + "'" )
	c.log.Info("PARAM: quantity = '" + quantity + "'")
	c.log.Info("PARAM: hasImg = '" + hasImg + "'")

	if name == ""{
		c.log.Error("[PARAM] received name is empty")
		c.AjaxSetResult(400, "param error")
		return
	}

	_, err := models.GetGoodsByName(name)
	if err != orm.ErrNoRows{
		c.log.Error("[PARAM] name is already exists")
		c.AjaxSetResult(400, "name is already exists")
		return
	}else if err != nil && err != orm.ErrNoRows {
		c.log.Error("[DB] database got an error: " + err.Error())
		c.AjaxSetResult(500, "database error: " + err.Error())
		return
	}

	pricef, err := strconv.ParseFloat(price, 10)
	if price == "" || err != nil {
		c.log.Error("[PARAM] received price is not a float")
		c.AjaxSetResult(400, "price is invalid")
		return
	}
	quantityi, err := strconv.ParseInt(quantity, 10, 64)
	if quantity == "" || err != nil {
		c.log.Error("[PARAM] received quantity is not an integer")
		c.AjaxSetResult(400, "quantity is invalid")
		return
	}

	if imageFile != nil{
		defer imageFile.Close()
		if imageHeader == nil{
			c.log.Error("imageHeader is nil when imageFile is not nil")
			c.AjaxSetResult(500, "imageHeader is nil when imageFile is not nil")
			return
		}
		sizeFormat, errFormat := common.NumberUnitFormat(imageHeader.Size, 2, 1024, 0, "")
		if errFormat != nil {
			sizeFormat = strconv.FormatInt(imageHeader.Size, 10)
		}
		c.log.Info("PARAM: image = ['" + imageHeader.Filename + "', '" + sizeFormat + "', '" + imageHeader.Header.Get(common.MINE_CONTENT_TYPE) + "'")
		imageSaveFilename, err = common.RerenderImage(imageFile, imageHeader)
		if err != nil{
			c.AjaxSetResult(400, "Image rerender failed: " + err.Error())
			return
		}
	}else if imageErr != nil{
		if hasImg == "ok"{
			c.AjaxSetResult(400, "Image was corrupted: " + imageErr.Error())
			return
		}
	}else{
		c.AjaxSetResult(400, "Image upload parse failed.")
		return
	}

	good := models.Good{
		Name: name,
		Desc: desc,
		Price: pricef,
		Quantity: quantityi,
	}

	if imageSaveFilename != ""{
		good.Image = imageSaveFilename
	}

	id, err := models.AddGoods(&good)
	if err != nil {
		c.log.Error("[AddGoods] Database error: " + err.Error())
		c.AjaxSetResult(500, "database error: " + err.Error())
		return
	}

	c.res.Data["id"] = id
	c.AjaxSetResult(200, "success")
	return

	//if name == "" {
	//	c.AjaxSetResult(400, "Expect an argument 'name', but not found.")
	//	return
	//}
	//if err != nil {
	//	c.AjaxSetResult(500, "Image not a vaild err: "+err.Error())
	//	return
	//}
	//imagePath := "static/upload/" + handle.Filename
	//c.SaveToFile("image", imagePath)
	//
	//var floatPrice float64
	//
	//var intQuantity int64
	//var id int64
	//if price == "" {
	//	floatPrice = 0
	//}
	//if floatPrice, err = strconv.ParseFloat(price, 64); err != nil {
	//	c.AjaxSetResult(400, "Expect an float type 'price', but parse failed: "+err.Error())
	//	return
	//}
	//if quantity == "" {
	//	intQuantity = 0
	//}
	//if intQuantity, err = strconv.ParseInt(quantity, 10, 64); err != nil {
	//	c.AjaxSetResult(400, "Expect an int type 'quantity', but parse failed: "+err.Error())
	//	return
	//}
	//good := models.Good{Name: name, Desc: desc, Price: floatPrice, Quantity: intQuantity}
	//if id, err = models.AddGoods(&good); err != nil {
	//	c.AjaxSetResult(501, "Cannot insert a good to database, system error: "+err.Error())
	//	return
	//}
	//fmt.Println("con: ", id, err == nil)
	//c.res.Data["id"] = id
	//c.AjaxSetResult(200, "Success")
}
