package controllers

import (
	"beego-goodsm/common"
	"beego-goodsm/models"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"strconv"
	"strings"
	"time"
)

type HistoryController struct{
	beego.Controller
	log *logs.BeeLogger
}

func (c *HistoryController) Prepare(){
	c.log = logs.GetBeeLogger()
}

func (c *HistoryController) Get(){
	var err error
	colors := map[string]string{"登记":"table-light", "售出":"table-success", "进货":"table-info", "撤柜":"table-warning", "删除":"table-danger", "恢复":"table-secondary", "更新":"table-primary"}
	name := c.Ctx.Input.Query("name")
	startTimeStr := c.Ctx.Input.Query("st")
	endTimeStr := c.Ctx.Input.Query("et")
	orderStr := c.Ctx.Input.Query("o")
	eventStr := c.Ctx.Input.Query("e")
	if name != ""{
		name = strings.Replace(name, "'", "\\'", -1)
	}
	c.Data["name"] = name

	now := time.Now()
	startTime, _ := common.GetTimeBeginning("day", &now)
	if startTimeStr != ""{
		startTime, err = time.Parse(common.WiredDate, startTimeStr)
		if err != nil {
			startTime, _ = common.GetTimeBeginning("day", &now)
		}
		startTime, _ = common.GetTimeBeginning("day", &startTime)
	}
	c.Data["startDate"] = startTime.Format(common.WiredDate)
	endTime, _ := common.GetTimeEnding("day", &now)
	if endTimeStr != ""{
		endTime, err = time.Parse(common.WiredDate, endTimeStr)
		if err != nil {
			endTime, _ = common.GetTimeEnding("day", &now)
		}
		endTime, _ = common.GetTimeEnding("day", &endTime)
	}
	c.Data["endDate"] = endTime.Format(common.WiredDate)

	order := models.ORDERBY_DESC
	if orderStr == "a"{
		order = models.ORDERBY_ASC
	}else if orderStr == "d"{
		order = models.ORDERBY_DESC
	}

	event := models.EVENT_SELL
	if eventStr == models.EVENT_IMPORT ||
		eventStr == models.EVENT_EXPORT ||
		eventStr == models.EVENT_ADD ||
		eventStr == models.EVENT_SELL ||
		eventStr == models.EVENT_DELETE ||
		eventStr == models.EVENT_RECOVER ||
		eventStr == models.EVENT_UPDATE ||
		eventStr == "all"{
		event = eventStr
	}
	c.Data["event"] = event
	if eventStr == "all"{
		event = ""
	}

	stat, err := models.StatEventSummary(&startTime, &endTime, name, event, models.STAT_SUM_QUANTITY | models.STAT_SUM_MONEY | models.STAT_SUM_PROFITS)
	if err != nil {
		c.Data["statErrorStr"] = "服务器出现错误: " + err.Error()
	}else{
		c.Data["statSumQuantity"] = strconv.FormatFloat(stat[models.STAT_SUM_QUANTITY_KEY], 'f', 0, 64)
		c.Data["statSumMoney"] = strconv.FormatFloat(stat[models.STAT_SUM_MONEY_KEY], 'f', 2, 64)
		c.Data["statSumProfits"] = strconv.FormatFloat(stat[models.STAT_SUM_PROFITS_KEY], 'f', 2, 64)
	}

	his, err := models.GoodHistoryByName(&startTime, &endTime, event, name, order)
	var history []map[string]interface{}
	if err != nil {
		c.Data["errorStr"] = "服务器出现错误: " + err.Error()
	}else{
		for _, h := range his{
			hh, err := common.Struct2Map(h, true)
			if err != nil{
				c.log.Warn("[Struct2Map] transform failed. error: " + err.Error() + ", data ignored.")
				continue
			}
			hh["createtime"] = h.CreateTime.Format(common.WiredTime)
			goodPrice, ok1 := hh["goodprice"].(float64)
			quantity, ok2 := hh["quantity"].(int64)
			money, ok3 := hh["money"].(float64)
			if !ok1 || !ok2 || !ok3{
				hh["profits"] = ""
			}else{
				hh["profits"] = strconv.FormatFloat(money - (goodPrice * float64(quantity)), 'f', 2, 64)
			}

			if event == ""{
				hh["color"] = colors[hh["event"].(string)]
			}
			history = append(history, hh)
		}
		c.Data["history"] = history
	}

	c.TplName = "history.html"
}