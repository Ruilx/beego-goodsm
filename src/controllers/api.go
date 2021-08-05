package controllers

import (
	"beego-goodsm/common"
	"beego-goodsm/models"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
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
	case "upd": // Update Good
		c.UpdGood()
	case "del": // Delete Good
		c.DeleteGood()
	case "sel": // Sell Good
		c.SellGood()
	case "get": // Get Goods
		c.GetGoods()
	case "imp": // Import Goods
		c.ImportGood()
	case "exp": // Export Goods
		c.ExportGood()
	case "rcv": // Recover Goods
		c.RecoverGood()

	case "statsel": //今日售出/本月售出
		c.StatSell()
	default:
		c.res.Msg = "Not a valid operation"
	}

	fmt.Println(c.res)
}

func (c *MainController) StatSell(){
	idJson := c.Ctx.Input.Query("ids")
	var ids []int64
	err := json.Unmarshal([]byte(idJson), &ids)
	if err != nil{
		c.AjaxSetResult(400, err.Error())
		return
	}
	year, month, day := time.Now().Date()
	endTime := time.Now()
	todayBeginning := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	thisMonthBeginning := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)

	resultToday, err := models.StatSellGoodsByGoodId(&todayBeginning, &endTime, ids, models.STAT_COUNT_ITEMS | models.STAT_SUM_QUANTITY | models.STAT_SUM_MONEY | models.STAT_SUM_PROFITS)
	if err != nil {
		c.AjaxSetResult(500, err.Error())
		return
	}
	resultThisMonth, err := models.StatSellGoodsByGoodId(&thisMonthBeginning, &endTime, ids, models.STAT_COUNT_ITEMS | models.STAT_SUM_QUANTITY | models.STAT_SUM_MONEY | models.STAT_SUM_PROFITS)
	if err != nil {
		c.AjaxSetResult(500, err.Error())
		return
	}

	result := make(map[int64]map[string]float64)
	for _, id := range ids{
		if goodsDStat, ok := resultToday[id]; ok{
			if statCountItems, ok := goodsDStat[models.STAT_COUNT_ITEMS_KEY]; ok{
				result[id][models.STAT_COUNT_ITEMS_KEY + "_day"] = statCountItems
			}else{
				result[id][models.STAT_COUNT_ITEMS_KEY + "_day"] = 0
			}
			if statSumQuantity, ok := goodsDStat[models.STAT_SUM_QUANTITY_KEY]; ok{
				result[id][models.STAT_SUM_QUANTITY_KEY + "_day"] = statSumQuantity
			}else{
				result[id][models.STAT_SUM_QUANTITY_KEY + "_day"] = 0
			}
			if statSumMoney, ok := goodsDStat[models.STAT_SUM_MONEY_KEY]; ok{
				result[id][models.STAT_SUM_MONEY_KEY + "_day"] = statSumMoney
			}else{
				result[id][models.STAT_SUM_MONEY_KEY + "_day"] = 0
			}
			if statSumProfits, ok := goodsDStat[models.STAT_SUM_PROFITS_KEY]; ok{
				result[id][models.STAT_SUM_PROFITS_KEY + "_day"] = statSumProfits
			}else{
				result[id][models.STAT_SUM_PROFITS_KEY + "_day"] = 0
			}
		}else{
			result[id][models.STAT_COUNT_ITEMS_KEY + "_day"] = 0
			result[id][models.STAT_SUM_QUANTITY_KEY + "_day"] = 0
			result[id][models.STAT_SUM_MONEY_KEY + "_day"] = 0
			result[id][models.STAT_SUM_PROFITS_KEY + "_day"] = 0
		}
		if goodsMStat, ok := resultThisMonth[id]; ok{
			if statCountItems, ok := goodsMStat[models.STAT_COUNT_ITEMS_KEY]; ok{
				result[id][models.STAT_COUNT_ITEMS_KEY + "_month"] = statCountItems
			}else{
				result[id][models.STAT_COUNT_ITEMS_KEY + "_month"] = 0
			}
			if statSumQuantity, ok := goodsMStat[models.STAT_SUM_QUANTITY_KEY]; ok{
				result[id][models.STAT_SUM_QUANTITY_KEY + "_month"] = statSumQuantity
			}else{
				result[id][models.STAT_SUM_QUANTITY_KEY + "_month"] = 0
			}
			if statSumMoney, ok := goodsMStat[models.STAT_SUM_MONEY_KEY]; ok{
				result[id][models.STAT_SUM_MONEY_KEY + "_month"] = statSumMoney
			}else{
				result[id][models.STAT_SUM_MONEY_KEY + "_month"] = 0
			}
			if statSumProfits, ok := goodsMStat[models.STAT_SUM_PROFITS_KEY]; ok{
				result[id][models.STAT_SUM_PROFITS_KEY + "_month"] = statSumProfits
			}else{
				result[id][models.STAT_SUM_PROFITS_KEY + "_month"] = 0
			}
		}else{
			result[id][models.STAT_COUNT_ITEMS_KEY + "_month"] = 0
			result[id][models.STAT_SUM_QUANTITY_KEY + "_month"] = 0
			result[id][models.STAT_SUM_MONEY_KEY + "_month"] = 0
			result[id][models.STAT_SUM_PROFITS_KEY + "_month"] = 0
		}
	}

	fmt.Println(result)

	c.AjaxSetResult(200, idJson)
}

func (c *MainController) GetGoods() {
	name := c.Ctx.Input.Query("name")

	fmt.Println("Search: " + name)

	var goods []models.Good
	var err error
	if goods, err = models.GetGoods(name, models.ORDERBY_ASC); err != nil {
		c.AjaxSetResult(500, err.Error())
		return
	}
	fmt.Println(goods)

	if name != "" {
		if len(goods) > 0 {
			var mp map[string]interface{}
			good := goods[0]
			if mp, err = common.Struct2Map(good, true); err != nil {
				c.AjaxSetResult(500, err.Error())
				return
			}
			result := make(map[string]interface{})
			var goodsMap []map[string]interface{}
			goodsMap = append(goodsMap, mp)
			result["goods"] = goodsMap
			c.res.Data = result
			c.AjaxSetResult(200, "success")
			return
		} else {
			result := make(map[string]interface{})
			goodsMap := make([]map[string]interface{}, 0)
			result["goods"] = goodsMap
			c.res.Data = result
			c.AjaxSetResult(200, "success, no result")
		}

	} else {
		var goodsResult []map[string]interface{}
		for _, v := range goods {
			var good map[string]interface{}
			if good, err = common.Struct2Map(v, true); err != nil {
				c.log.Warn("One good parse failed: " + v.Name + ", error: " + err.Error())
				continue
			}
			goodsResult = append(goodsResult, good)
		}
		result := make(map[string]interface{})
		result["goods"] = goodsResult
		c.res.Data = result
		c.AjaxSetResult(200, "success")
		return
	}

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
	c.log.Info("PARAM: price = '" + price + "'")
	c.log.Info("PARAM: quantity = '" + quantity + "'")
	c.log.Info("PARAM: hasImg = '" + hasImg + "'")

	if name == "" {
		c.log.Error("[PARAM] received name is empty")
		c.AjaxSetResult(400, "param error")
		return
	}

	_, err := models.GetGoodsByName(name)
	if err != orm.ErrNoRows {
		c.log.Error("[PARAM] name is already exists")
		c.AjaxSetResult(400, "name is already exists")
		return
	} else if err != nil && err != orm.ErrNoRows {
		c.log.Error("[DB] database got an error: " + err.Error())
		c.AjaxSetResult(500, "database error: "+err.Error())
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

	if imageFile != nil {
		defer imageFile.Close()
		if imageHeader == nil {
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
		if err != nil {
			c.AjaxSetResult(400, "Image rerender failed: "+err.Error())
			return
		}
	} else if imageErr != nil {
		if hasImg == "ok" {
			c.AjaxSetResult(400, "Image was corrupted: "+imageErr.Error())
			return
		}
	} else {
		c.AjaxSetResult(400, "Image upload parse failed.")
		return
	}

	good := models.Good{
		Name:     name,
		Desc:     desc,
		Price:    pricef,
		Quantity: quantityi,
	}

	if imageSaveFilename != "" {
		good.Image = imageSaveFilename
	}

	var resultId int64
	var resultHisId int64

	resultId, err = models.AddGoods(&good)
	if err != nil {
		c.log.Error("[AddGoods] Database error: " + err.Error())
		c.AjaxSetResult(500, "database error: "+err.Error())
		return
	}

	// writing sell history
	tryTimes := 5
	i := 0
	for i = 0; i < tryTimes; i++ {
		if resultHisId, err = models.AddAddHistory(&good, "新货品登记"); err == nil {
			break
		}
	}
	if i >= tryTimes && err != nil {
		c.log.Error("cannot insert good history, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot insert good history, db error: "+err.Error())
		return
	}

	c.res.Data["result_id"] = resultId
	c.res.Data["result_his_id"] = resultHisId
	c.AjaxSetResult(200, "success")
	return

}

func (c *MainController) UpdGood() {
	id := c.Ctx.Input.Query("id")
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
	c.log.Info("PARAM: price = '" + price + "'")
	c.log.Info("PARAM: quantity = '" + quantity + "'")
	c.log.Info("PARAM: hasImg = '" + hasImg + "'")

	var idi int
	var err error
	if id == "" {
		c.log.Error("[PARAM] received id is empty")
		c.AjaxSetResult(400, "param error")
		return
	} else if idi, err = strconv.Atoi(id); err != nil {
		c.log.Error("[PARAM] received id is not an integer: id: '" + id + "', error: " + err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}
	if name == "" {
		c.log.Error("[PARAM] received name is empty")
		c.AjaxSetResult(400, "param error")
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

	if imageFile != nil {
		defer imageFile.Close()
		if imageHeader == nil {
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
		if err != nil {
			c.AjaxSetResult(400, "Image rerender failed: "+err.Error())
			return
		}
	} else if imageErr != nil {
		if hasImg == "ok" {
			c.AjaxSetResult(400, "Image was corrupted: "+imageErr.Error())
			return
		}
	} else {
		c.AjaxSetResult(400, "Image upload parse failed.")
		return
	}

	good := models.Good{
		Id:       int32(idi),
		Name:     name,
		Desc:     desc,
		Price:    pricef,
		Quantity: quantityi,
	}

	if imageSaveFilename != "" {
		good.Image = imageSaveFilename
	}

	insertId, err := models.UpdateGoodsById(&good)
	if err != nil {
		c.log.Error("[AddGoods] Database error: " + err.Error())
		c.AjaxSetResult(500, "database error: "+err.Error())
		return
	}

	c.res.Data["id"] = insertId
	c.AjaxSetResult(200, "success")
	return
}

func (c *MainController) SellGood() {
	id := c.Ctx.Input.Query("id")                // ID
	quantity := c.Ctx.Input.Query("quantity")    // sell quantity
	price := c.Ctx.Input.Query("price")          // sell price
	remark := c.Ctx.Input.Query("remark")        // remark
	unitPrice := c.Ctx.Input.Query("unit_price") // unit price

	var err error
	var idi int
	var quantityi int64
	var pricef float64
	var unitPricef float64

	c.log.Info("Enrty SellGood operation.")
	c.log.Info("[PARAM] id = ", id)
	c.log.Info("[PARAM] quantity = ", quantity)
	c.log.Info("[PARAM] price = ", price)
	c.log.Info("[PARAM] remark = ", remark)
	c.log.Info("[PARAM] unitPrice = ", unitPrice)

	if idi, err = strconv.Atoi(id); err != nil {
		c.log.Error("cannot parse id to int: ", err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}
	if idi <= 0 {
		c.log.Error("not an valid id: ", idi)
		c.AjaxSetResult(400, "param error")
		return
	}
	if quantityi, err = strconv.ParseInt(quantity, 10, 64); err != nil {
		c.log.Error("cannot parse quantity to int: ", err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}
	if pricef, err = strconv.ParseFloat(price, 64); err != nil {
		c.log.Error("cannot parse price to float64: ", err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}
	if unitPricef, err = strconv.ParseFloat(unitPrice, 64); err != nil {
		c.log.Error("cannot parse unit_price to float64: ", err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}

	thisGood, err := models.GetGoodsById(int32(idi))
	if err != nil {
		c.log.Error("[Sell Good] db error: ", err.Error())
		c.AjaxSetResult(500, "database error: "+err.Error())
		return
	}

	if thisGood.Price != unitPricef {
		c.log.Error("sent request unit_price is not equal to current price.")
		c.log.Error("CurrentUnitPrice: ", thisGood.Price, "; sent Price: ", unitPricef)
		c.AjaxSetResult(400, "the current unit price is changed recently, please resend this request.")
		return
	}
	if pricef == 0.0 {
		c.log.Info("[Sell Good] Sell good for free:" + thisGood.Name + " x" + quantity)
	}
	if thisGood.Quantity <= 0 {
		c.log.Error("cannot sell goods while has no or negative storage: ", thisGood.Quantity)
		c.AjaxSetResult(403, "cannot sell goods while has no or negative storage: "+strconv.FormatInt(thisGood.Quantity, 10))
		return
	}

	balance := thisGood.Quantity - quantityi
	if balance < 0 {
		c.log.Info("[Sell Good] sold to negative balance")
	}

	balancedGood := models.Good{
		Id:       thisGood.Id,
		Quantity: balance,
	}

	var resultId int64
	var resultHisId int64
	if resultId, err = models.UpdateGoodsById(&balancedGood, "Quantity"); err != nil {
		c.log.Error("cannot update good balance, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot update good balance, db error: "+err.Error())
		return
	}

	// writing sell history
	tryTimes := 5
	i := 0
	for i = 0; i < tryTimes; i++ {
		if resultHisId, err = models.AddSellsHistory(thisGood, quantityi, unitPricef, float64(quantityi)*unitPricef, remark, balance); err == nil {
			break
		}
	}
	if i >= tryTimes && err != nil {
		c.log.Error("cannot insert good history, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot insert good history, db error: "+err.Error())
		return
	}

	c.res.Data["result_id"] = resultId
	c.res.Data["result_his_id"] = resultHisId
	c.AjaxSetResult(200, "success")
	return
}

func (c *MainController) ImportGood() {
	id := c.Ctx.Input.Query("id")             // ID
	quantity := c.Ctx.Input.Query("quantity") // quantity
	remark := c.Ctx.Input.Query("remark")

	var err error
	var idi int64
	var quantityi int64

	c.log.Info("Entry ImportGood operation.")
	c.log.Info("[PARAM] id = ", id)
	c.log.Info("[PARAM] quantity = ", quantity)
	c.log.Info("[PARAM] remark = ", remark)

	if idi, err = strconv.ParseInt(id, 10, 32); err != nil {
		c.log.Error("cannot parse id to int: ", err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}
	if idi <= 0 {
		c.log.Error("not an valid id: ", idi)
		c.AjaxSetResult(400, "param error")
		return
	}
	if quantityi, err = strconv.ParseInt(quantity, 10, 64); err != nil {
		c.log.Error("cannot parse quantity to int: ", err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}
	if quantityi <= 0 {
		c.log.Error("not an valid quntity: ", quantity)
		c.AjaxSetResult(400, "param error")
		return
	}

	thisGood, err := models.GetGoodsById(int32(idi))
	if err != nil {
		c.log.Error("[Import Good] db error: ", err.Error())
		c.AjaxSetResult(500, "database error")
		return
	}

	newBalance := thisGood.Quantity + quantityi
	if newBalance < thisGood.Quantity {
		c.log.Error("[Import Good] param quantity overflowed. Current quantity: ", newBalance)
		c.AjaxSetResult(403, "param quantity maybe to large cannot be import.")
		return
	}
	balancedGood := models.Good{
		Id:       thisGood.Id,
		Quantity: newBalance,
	}

	var resultId int64
	var resultHisId int64
	if resultId, err = models.UpdateGoodsById(&balancedGood, "Quantity"); err != nil {
		c.log.Error("cannot update good balance, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot update good balance, db error: "+err.Error())
		return
	}

	// writing import history
	tryTimes := 5
	i := 0
	for i = 0; i < tryTimes; i++ {
		if resultHisId, err = models.AddImportHistory(thisGood, quantityi, remark, newBalance); err == nil {
			break
		}
	}
	if i >= tryTimes && err != nil {
		c.log.Error("cannot insert good history, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot insert good history, db error: "+err.Error())
		return
	}
	c.res.Data["result_id"] = resultId
	c.res.Data["result_his_id"] = resultHisId
	c.AjaxSetResult(200, "success")
	return
}

func (c *MainController) ExportGood() {
	id := c.Ctx.Input.Query("id")
	quantity := c.Ctx.Input.Query("quantity")
	remark := c.Ctx.Input.Query("remark")

	var err error
	var idi int64
	var quantityi int64

	c.log.Info("Entry ExportGood operation.")
	c.log.Info("[PARAM] id = ", id)
	c.log.Info("[PARAM] quantity = ", quantity)
	c.log.Info("[PARAM] remark = ", remark)

	if idi, err = strconv.ParseInt(id, 10, 32); err != nil {
		c.log.Error("cannot parse id to int: ", err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}
	if idi <= 0 {
		c.log.Error("not an vaild id: ", idi)
		c.AjaxSetResult(400, "param error")
		return
	}
	if quantityi, err = strconv.ParseInt(quantity, 10, 64); err != nil {
		c.log.Error("cannot parse quantity to int: ", err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}
	if quantityi <= 0 {
		c.log.Error("not an valid quantity: ", quantity)
		c.AjaxSetResult(400, "param error")
		return
	}

	thisGood, err := models.GetGoodsById(int32(idi))
	if err != nil {
		c.log.Error("[Export Good] db error: ", err.Error())
		c.AjaxSetResult(500, "database error")
		return
	}

	newBalance := thisGood.Quantity - quantityi
	if newBalance > thisGood.Quantity {
		c.log.Error("[Export Good] param quantity overflowed. Current quantity: ", newBalance)
		c.AjaxSetResult(403, "param quantity maybe to large cannot be export.")
		return
	}
	balancedGood := models.Good{
		Id:       thisGood.Id,
		Quantity: newBalance,
	}

	var resultId int64
	var resultHisId int64
	if resultId, err = models.UpdateGoodsById(&balancedGood, "Quantity"); err != nil {
		c.log.Error("cannot update good balance, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot update good balance, db error: "+err.Error())
		return
	}

	// writing export history
	tryTimes := 5
	i := 0
	for i = 0; i < tryTimes; i++ {
		if resultHisId, err = models.AddExportHistory(thisGood, quantityi, remark, newBalance); err == nil {
			break
		}
	}
	if i >= tryTimes && err != nil {
		c.log.Error("cannot insert good history, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot insert good history, db error: "+err.Error())
		return
	}
	c.res.Data["result_id"] = resultId
	c.res.Data["result_his_id"] = resultHisId
	c.AjaxSetResult(200, "success")
	return
}

func (c *MainController) DeleteGood() {
	id := c.Ctx.Input.Query("id")
	remark := c.Ctx.Input.Query("remark")

	var err error
	var idi int64

	c.log.Info("Entry DeleteGood operation.")
	c.log.Info("[PARAM] id: ", idi)
	c.log.Info("[PARAM] remark: ", remark)

	if idi, err = strconv.ParseInt(id, 10, 32); err != nil {
		c.log.Error("cannot parse id to int: ", err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}
	if idi <= 0 {
		c.log.Error("not an valid id: ", idi)
		c.AjaxSetResult(400, "param error")
		return
	}

	thisGood, err := models.GetGoodsById(int32(idi))
	if err != nil {
		c.log.Error("[Delete Good] db error: ", err.Error())
		c.AjaxSetResult(500, "database error")
		return
	}

	if thisGood.Deleted {
		c.log.Error("[Delete Good] the good was deleted.")
		c.AjaxSetResult(400, "Good: '"+thisGood.Name+"' was deleted.")
		return
	}

	balancedGood := models.Good{
		Id:         thisGood.Id,
		Deleted:    true,
		DeleteTime: time.Now(),
	}

	var resultId int64
	var resultHisId int64
	if resultId, err = models.UpdateGoodsById(&balancedGood, "Deleted", "DeleteTime"); err != nil {
		c.log.Error("cannot update good deleting status, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot update good deleting status, db error: "+err.Error())
		return
	}

	// writing deleted history
	tryTimes := 5
	i := 0
	for i = 0; i < tryTimes; i++ {
		if resultHisId, err = models.AddDeleteHistory(thisGood, remark); err == nil {
			break
		}
	}
	if i >= tryTimes && err != nil {
		c.log.Error("cannot update deleting history, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot udpate deleting history, db error: "+err.Error())
		return
	}
	c.res.Data["result_id"] = resultId
	c.res.Data["result_his_id"] = resultHisId
	c.AjaxSetResult(200, "success")
	return
}

func (c *MainController) RecoverGood() {
	id := c.Ctx.Input.Query("id")
	remark := c.Ctx.Input.Query("remark")

	var err error
	var idi int64

	c.log.Info("Entry RecoverGood operation.")
	c.log.Info("[PARAM] id: ", id)
	c.log.Info("[PARAM] remark: ", remark)

	if idi, err = strconv.ParseInt(id, 10, 32); err != nil {
		c.log.Error("cannot parse id to int: ", err.Error())
		c.AjaxSetResult(400, "param error")
		return
	}
	if idi <= 0 {
		c.log.Error("not an valid id: ", idi)
		c.AjaxSetResult(400, "param error")
		return
	}

	thisGood, err := models.GetGoodsById(int32(idi))
	if err != nil {
		c.log.Error("[Recover Good] db error: ", err.Error())
		c.AjaxSetResult(500, "database error")
		return
	}

	if !thisGood.Deleted {
		c.log.Error("[Recover Good] the good isn't deleted.")
		c.AjaxSetResult(400, "Good: '"+thisGood.Name+"' isn't deleted.")
		return
	}

	balancedGood := models.Good{
		Id:      thisGood.Id,
		Deleted: false,
	}

	var resultId int64
	var resultHisId int64
	if resultId, err = models.UpdateGoodsById(&balancedGood, "Deleted"); err != nil {
		c.log.Error("cannot update good deleting status, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot update good deleting status, db error: "+err.Error())
		return
	}

	// writing recovered history
	tryTimes := 5
	i := 0
	for i = 0; i < tryTimes; i++ {
		if resultHisId, err = models.AddRecoverHistory(thisGood, remark); err == nil {
			break
		}
	}
	if i >= tryTimes && err != nil {
		c.log.Error("cannot update recovering history, db error: " + err.Error())
		c.AjaxSetResult(500, "cannot update recovering history, db error: "+err.Error())
		return
	}
	c.res.Data["result_id"] = resultId
	c.res.Data["result_his_id"] = resultHisId
	c.AjaxSetResult(200, "success")
	return
}
