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
	c.Data["name"] = name

	ml, err := models.GetGoods(name, models.ORDERBY_DESC, false)
	if err != nil{
		c.log.Error("[DB] database error: " + err.Error())
		c.Data["err"] = "系统错误, 请检查系统日志文件输出"
		return
	}

	totalGoodsCount, totalGoodsQuantity, err := models.GetGoodsCount("")
	if err != nil{
		totalGoodsCount = 0
		totalGoodsQuantity = 0
	}
	c.Data["totalGoodsCount"] = totalGoodsCount
	c.Data["totalGoodsQuantity"] = totalGoodsQuantity

	c.Data["dayStatOk"] = false
	c.Data["daySumQuantity"] = 0
	c.Data["daySumMoney"] = 0
	c.Data["daySumProfits"] = 0
	c.Data["monthStatOk"] = false
	c.Data["monthSumQuantity"] = 0
	c.Data["monthSumMoney"] = 0
	c.Data["monthSumProfits"] = 0

	now := time.Now()
	c.Data["now"] = now.Format(common.WiredDate)

	dayStart, err := common.GetTimeBeginning("day", &now)
	if err == nil{
		dayStat, err := models.StatSoldSummary(&dayStart, &now, models.STAT_SUM_QUANTITY | models.STAT_SUM_MONEY | models.STAT_SUM_PROFITS)
		if err == nil {
			c.Data["dayStatOk"] = true
			c.Data["daySumQuantity"] = dayStat[models.STAT_SUM_QUANTITY_KEY]
			c.Data["daySumMoney"] = dayStat[models.STAT_SUM_MONEY_KEY]
			c.Data["daySumProfits"] = dayStat[models.STAT_SUM_PROFITS_KEY]
		}
	}
	monthStart, err := common.GetTimeBeginning("month", &now)
	if err == nil{
		monthStat, err := models.StatSoldSummary(&monthStart, &now, models.STAT_SUM_QUANTITY | models.STAT_SUM_MONEY | models.STAT_SUM_PROFITS)
		if err == nil {
			c.Data["monthStatOk"] = true
			c.Data["monthSumQuantity"] = monthStat[models.STAT_SUM_QUANTITY_KEY]
			c.Data["monthSumMoney"] = monthStat[models.STAT_SUM_MONEY_KEY]
			c.Data["monthSumProfits"] = monthStat[models.STAT_SUM_PROFITS_KEY]
		}
	}

	c.Data["isMobile"] = false
	ua := c.Ctx.Input.Header("User-Agent")
	if common.IsMobileUsingUserAgent(ua){
		c.Data["isMobile"] = true
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

	var ids = make([]int64, 0, len(goods))
	for v := range goods{
		ids = append(ids, goods[v]["id"].(int64))
	}

	dayGoodIdsStat, err := models.StatSellGoodsByGoodId(&dayStart, &now, ids, models.STAT_SUM_QUANTITY)
	if err == nil{
		for id, stat := range dayGoodIdsStat{
			var goodStat *map[string]interface{}
			for i := range goods{
				if goods[i]["id"] == id {
					goodStat = &goods[i]
					break
				}
			}
			if goodStat != nil{
				(*goodStat)["daySellCount"] = stat[models.STAT_SUM_QUANTITY_KEY]
			}else{
				c.log.Error("passed id '%d' not contains in goods list", id)
			}
		}
	}else{
		idsStr := make([]string, 0, len(ids))
		for id := range ids{
			idsStr = append(idsStr, strconv.FormatInt(ids[id], 10))
		}
		c.log.Error("Cannot get sold daily stat from history when id in ('" + strings.Join(idsStr, "', '") + "')")
	}

	monthGoodIdsStat, err := models.StatSellGoodsByGoodId(&monthStart, &now, ids, models.STAT_SUM_QUANTITY)
	if err == nil{
		for id, stat := range monthGoodIdsStat{
			var goodStat *map[string]interface{}
			for i := range goods{
				if goods[i]["id"] == id {
					goodStat = &goods[i]
					break
				}
			}
			if goodStat != nil{
				(*goodStat)["monthSellCount"] = stat[models.STAT_SUM_QUANTITY_KEY]
			}else{
				c.log.Error("passed id '%d' not contains in goods list", id)
			}
		}
	}else{
		idsStr := make([]string, 0, len(ids))
		for id := range ids{
			idsStr = append(idsStr, strconv.FormatInt(ids[id], 10))
		}
		c.log.Error("Cannot get monthly sold stat from history when id in ('" + strings.Join(idsStr, "', '") + "')")
	}

	// 将还没来得及有卖出记录的货品统计信息记录为0
	for i := range goods{
		stat := &goods[i]
		if _, ok := (*stat)["daySellCount"]; !ok{
			(*stat)["daySellCount"] = 0
		}
		if _, ok := (*stat)["monthSellCount"]; !ok{
			(*stat)["monthSellCount"] = 0
		}
	}
	c.Data["goods"] = goods

	c.TplName = "index.html"
}
