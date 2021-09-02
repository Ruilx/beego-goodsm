package models

import (
	"beego-goodsm/common"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

const DBNAME = "default"

const (
	ORDERBY_UNKNOWN = iota
	ORDERBY_ASC
	ORDERBY_DESC
)

const (
	STATUS_UNKNWON = iota // 不使用
	STATUS_ACTIVE         // 默认, 状态有效
	STATUS_DELETED        // 状态被人工删除, 但未还原
	STATUS_REVOKE         // 状态被撤回(比如撤回卖出等)
)

const (
	EVENT_ADD     = "登记"
	EVENT_SELL    = "售出"
	EVENT_IMPORT  = "进货"
	EVENT_EXPORT  = "撤柜"
	EVENT_DELETE  = "删除"
	EVENT_RECOVER = "恢复"
	EVENT_UPDATE  = "更新"
)

const (
	STAT_SUM_MONEY    = 0x00000001
	STAT_SUM_QUANTITY = 0x00000002
	STAT_SUM_PROFITS  = 0x00000004
	STAT_COUNT_ITEMS  = 0x00000008
)

const (
	STAT_SUM_MONEY_KEY    = "sum_money"
	STAT_SUM_QUANTITY_KEY = "sum_quantity"
	STAT_SUM_PROFITS_KEY  = "sum_profits"
	STAT_COUNT_ITEMS_KEY  = "count_items"
)

const (
	ORDERBY_ARGUMENT_ERROR = "error for order, needs one of ORDERBY_ASC, ORDERBY_DESC, ORDERBY_UNKNOWN"
)

type Good struct {
	Id         int32     `orm:"column(id); auto; pk"`
	Name       string    `orm:"column(name); size(32); unique"`
	Desc       string    `orm:"column(desc); type(text); null"`
	Price      float64   `orm:"column(price); default(0)"`
	Quantity   int64     `orm:"column(quantity); default(0)"`
	Image      string    `orm:"column(image); size(64); default()"`
	Deleted    bool      `orm:"column(deleted); default(0)"`
	CreateTime time.Time `orm:"column(create_time); auto_now_add; type(datetime)"`
	UpdateTime time.Time `orm:"column(update_time); auto_now; type(datetime)"`
	DeleteTime time.Time `orm:"column(delete_time); type(datetime); null; default(null)"`
}

type History struct {
	Id         int64     `orm:"column(id); auto; pk"`                              // 记录ID
	Event      string    `orm:"column(event); size(64);"`                          // 记录事件名
	GoodId     int32     `orm:"column(good_id)"`                                   // 货品ID
	GoodName   string    `orm:"column(good_name); size(32)"`                       // 货品名称快照
	GoodDesc   string    `orm:"column(good_desc); type(text); null"`               // 货品描述快照
	GoodPrice  float64   `orm:"column(good_price); default(0)"`                    // 货品价格快照
	GoodImage  string    `orm:"column(good_image); default()"`                     // 货品图片快照
	Quantity   int64     `orm:"column(quantity);"`                                 // 数量
	Money      float64   `orm:"column(money);"`                                    // 金额
	Remark     string    `orm:"column(remark); type(text); null"`                  // 备注
	Info       string    `orm:"column(info); default()"`                           // 信息
	Status     int8      `orm:"column(status); default(1)"`                        // 状态
	CreateTime time.Time `orm:"column(create_time); auto_now_add; type(datetime)"` // 创建时间
}

func init() {
	var (
		dbPath string
		err    error
	)

	orm.Debug = false

	if dbPath, err = web.AppConfig.String("db_path"); err != nil {
		panic("Cannot find 'dbPath' in app config: " + err.Error())
	}
	fmt.Println("DBPATH: " + dbPath)

	if err = orm.RegisterDriver("sqlite3", orm.DRSqlite); err != nil {
		panic("Cannot register sqlite3 for model: " + err.Error())
	}

	if err = orm.RegisterDataBase(DBNAME, "sqlite3", dbPath); err != nil {
		panic("Cannot register database " + DBNAME + " for sqlite3: " + err.Error())
	}

	if err = orm.SetDataBaseTZ(DBNAME, time.Local); err != nil {
		panic("Cannot set database " + DBNAME + " timezone to local for sqlite3: " + err.Error())
	}

	// 注册表
	orm.RegisterModel(&Good{})
	orm.RegisterModel(&History{})
	if err = orm.RunSyncdb(DBNAME, false, true); err != nil {
		panic("Run sync db failed, error: " + err.Error())
	}
}

// 添加货物
// 输入货品结构体
// 返回插入id和err
func AddGoods(goods *Good) (id int64, err error) {
	fmt.Println("Good: ", goods)
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	id, err = ormHandle.Insert(goods)

	return
}

// 按照ID获取货物信息
// 输入货物ID
// 返回货物结构体和err
func GetGoodsById(id int32) (res *Good, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	result := &Good{Id: id}
	if err = ormHandle.QueryTable(&Good{}).Filter("Id", id).RelatedSel().One(result); err != nil {
		return nil, err
	}
	return result, err
}

// 按照名称获取货物信息
// 输入货物名称(精确)
// 返回货物结构体和err
func GetGoodsByName(name string) (res *Good, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	result := &Good{Name: name}
	if err = ormHandle.QueryTable(&Good{}).Filter("Name", name).RelatedSel().One(result); err != nil {
		return nil, err
	}
	return result, err
}

// 通过名字筛选货品
// 输入货品名称(关键字)
// 返回货物结构体列表和err
func GetGoods(name string, order int, exact bool) (ml []Good, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	queryCond := ormHandle.QueryTable(&Good{})

	if name != "" {
		if exact {
			queryCond = queryCond.Filter("name", name)
		} else {
			queryCond = queryCond.Filter("name__icontains", name)
		}
	}
	queryCond = queryCond.Filter("deleted", "0")

	if order == ORDERBY_UNKNOWN || order == ORDERBY_ASC {
		queryCond = queryCond.OrderBy("create_time")
	} else {
		queryCond = queryCond.OrderBy("-create_time")
	}

	if _, err = queryCond.All(&ml); err != nil {
		return nil, err
	}
	return ml, nil
}

func GetGoodsCount(name string)(count int64, quantity int64, err error){
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	sql := "Select {SELECT} from good where {WHERE}"
	sel := make([]string, 0, 2)
	whe := make([]string, 0, 2)

	sel = append(sel, "count(1) as count")
	sel = append(sel,  "sum(quantity) as quantity")
	if name = strings.Replace(name, "'", "\\'", -1); name != ""{
		whe = append(whe, "name like '%" + name + "%'")
	}
	whe = append(whe, "deleted = 0")

	sql = strings.Replace(sql, "{SELECT}", strings.Join(sel, ","), 1)
	sql = strings.Replace(sql, "{WHERE}", strings.Join(whe, " and "), 1)

	resultInterface := make([]orm.Params, 0)
	_, err = ormHandle.Raw(sql).Values(&resultInterface)

	if err != nil{
		return
	}

	if len(resultInterface) <= 0 {
		fmt.Println("GetGoodsCount: result is nil")
		return 0, 0,nil
	}

	countRow := &resultInterface[0]
	ok := false
	var countStr string
	if countStr, ok = (*countRow)["count"].(string); !ok{
		fmt.Println("GetGoodsCount: result has no item named 'count'")
		return 0, 0,nil
	}
	if count, err = strconv.ParseInt(countStr, 10, 64); err != nil{
		fmt.Println("GetGoodsCount: result 'count' cannot parse to int64")
		return 0, 0, nil
	}
	var quantityStr string
	if quantityStr, ok = (*countRow)["quantity"].(string); !ok{
		fmt.Println("GetGoodsCount: result has no item named 'quantity'")
		return count, 0, nil
	}
	if quantity, err = strconv.ParseInt(quantityStr, 10, 64); err != nil{
		fmt.Println("GetGoodsCount: result 'quantity' cannot parse to int64")
		return count, 0, nil
	}
	return count, quantity,nil
}

// 获得所有货物
// (不常用)
func GetGoods2(query map[string]string, fields []string, sortBy []string, order []int, offset int64, limit int64) (ml []interface{}, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	queryCond := ormHandle.QueryTable(&Good{})

	for k, v := range query {
		k = strings.Replace(k, ".", "__", -1)
		queryCond = queryCond.Filter(k, v)
	}

	var sortFields []string
	if len(sortBy) != 0 {
		if len(sortBy) == len(order) {
			for i, v := range sortBy {
				orderBy := ""
				if order[i] == ORDERBY_DESC {
					orderBy = "-" + v
				} else if order[i] == ORDERBY_ASC {
					orderBy = v
				} else {
					return nil, errors.New("Error for order argument, value must be in ORDERBY_ASC or ORDERBY_DESC")
				}
				sortFields = append(sortFields, orderBy)
			}
		} else if len(sortBy) != len(order) && len(order) == 1 {
			for _, v := range sortBy {
				orderBy := ""
				if order[0] == ORDERBY_DESC {
					orderBy = "-" + v
				} else if order[0] == ORDERBY_ASC {
					orderBy = v
				} else {
					return nil, errors.New("Error for order argument, value must be in ORDERBY_ASC or ORDERBY_DESC")
				}
				sortFields = append(sortFields, orderBy)
			}
		} else if len(sortBy) != len(order) && len(order) != 1 {
			return nil, errors.New("Error for order and sortby, length mismatched")
		}
	} else {
		if len(order) != 0 {
			return nil, errors.New("Error, unused 'order' fields")
		}
	}
	var l []Good
	queryCond = queryCond.OrderBy(sortFields...).RelatedSel()
	if _, err = queryCond.Limit(limit, offset).All(&l, fields...); err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		for _, v := range l {
			ml = append(ml, v)
		}
	} else {
		for _, v := range l {
			m := make(map[string]interface{})
			val := reflect.ValueOf(v)
			for _, fname := range fields {
				m[fname] = val.FieldByName(fname).Interface()
			}
			ml = append(ml, m)
		}
	}
	return ml, nil
}

// 按照ID更新货物 ID写至good.Id
// 输入货品新的内容, 以及各种列名
// 返回更新id和err
func UpdateGoodsById(good *Good, col ...string) (id int64, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	idReady := Good{Id: good.Id}
	if err = ormHandle.Read(&idReady); err == nil {
		return ormHandle.Update(good, col...)
	}
	return
}

// 按照ID删除货物
// 输入删除货品的id
// 返回删除的dbId和err
func DeleteGoodsById(id int32) (dbId int64, err error) {
	ormHandle := orm.NewOrm()
	good := Good{Id: id}
	if err = ormHandle.Read(&good); err == nil {
		return ormHandle.Delete(&good)
	}
	return
}

// 写入售货历史表
// 输入新增的history
// 返回id和err
func AddHistory(his *History) (id int64, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	id, err = ormHandle.Insert(his)
	return
}

// 写入增加货品历史记录
// 输入: 货品结构体, 备注
// 返回: id和err
func AddAddHistory(good *Good, remark string) (id int64, err error){
	his := History{
		Event:     EVENT_ADD,
		GoodId:    good.Id,
		GoodName:  good.Name,
		GoodDesc:  good.Desc,
		GoodPrice: good.Price,
		GoodImage: good.Image,
		Quantity:  good.Quantity,
		Money:     good.Price * float64(good.Quantity),
		Remark:    remark,
		Info:      "[登记]: 登记【" + good.Name + "】，数量【" + strconv.FormatInt(good.Quantity, 10) + "】，单价【" + strconv.FormatFloat(good.Price, 'f', 2, 64) + "】，总价【" + strconv.FormatFloat(good.Price * float64(good.Quantity), 'f', 2, 64) + "】。",
		Status:    1,
	}
	return AddHistory(&his)
}

// 写入售出历史记录
// 输入: 售出货物快照, 数量, 单价, 总价, 备注, 余量
// 返回: id和err
func AddSellsHistory(good *Good, quantity int64, unitPrice float64, money float64, remark string, balance int64) (id int64, err error) {
	his := History{
		Event:     EVENT_SELL,
		GoodId:    good.Id,
		GoodName:  good.Name,
		GoodDesc:  good.Desc,
		GoodPrice: good.Price,
		GoodImage: good.Image,
		Quantity:  quantity,
		Money:     money,
		Remark:    remark,
		Info:      "[出售]: 卖出【" + good.Name + "】，数量【" + strconv.FormatInt(quantity, 10) + "】，单价【" + strconv.FormatFloat(unitPrice, 'f', 2, 64) + "】，总价【" + strconv.FormatFloat(money, 'f', 2, 64) + "】，剩余库存【" + strconv.FormatInt(balance, 10) + "】。",
		Status:    1,
	}
	return AddHistory(&his)
}

// 写入进货历史记录
// 输入: 进货货品快照, 数量, 备注, 余量
// 返回: id和err
func AddImportHistory(good *Good, quantity int64, remark string, balance int64) (id int64, err error) {
	money := good.Price * float64(quantity) // 进货物品所产生的定价价值
	his := History{
		Event:     EVENT_IMPORT,
		GoodId:    good.Id,
		GoodName:  good.Name,
		GoodDesc:  good.Desc,
		GoodPrice: good.Price,
		GoodImage: good.Image,
		Quantity:  quantity,
		Money:     money,
		Remark:    remark,
		Info:      "[进货]: 进货【" + good.Name + "】，数量【" + strconv.FormatInt(quantity, 10) + "】，库存【" + strconv.FormatInt(good.Quantity, 10) + "】，总价【" + strconv.FormatFloat(money, 'f', 2, 64) + "】，当前库存【" + strconv.FormatInt(balance, 10) + "】。",
		Status:    1,
	}
	return AddHistory(&his)
}

// 写入撤柜历史记录
// 输入: 撤柜货品快照, 数量, 备注, 余量
// 返回: id和err
func AddExportHistory(good *Good, quantity int64, remark string, balance int64) (id int64, err error) {
	money := good.Price * float64(quantity)
	his := History{
		Event:     EVENT_EXPORT,
		GoodId:    good.Id,
		GoodName:  good.Name,
		GoodDesc:  good.Desc,
		GoodPrice: good.Price,
		GoodImage: good.Image,
		Quantity:  quantity,
		Money:     money,
		Remark:    remark,
		Info:      "[撤柜]: 撤柜【" + good.Name + "】，数量【" + strconv.FormatInt(quantity, 10) + "】，库存【" + strconv.FormatInt(good.Quantity, 10) + "】，总价【" + strconv.FormatFloat(money, 'f', 2, 64) + "】，当前库存【" + strconv.FormatInt(balance, 10) + "】。",
		Status:    1,
	}
	return AddHistory(&his)
}

// 写入删除货品历史记录
// 输入: 删除货品快照, 备注
// 返回: id和err
func AddDeleteHistory(good *Good, remark string) (id int64, err error) {
	his := History{
		Event:     EVENT_DELETE,
		GoodId:    good.Id,
		GoodName:  good.Name,
		GoodDesc:  good.Desc,
		GoodPrice: good.Price,
		GoodImage: good.Image,
		Quantity:  good.Quantity,
		Money:     float64(good.Quantity) * good.Price,
		Remark:    remark,
		Info:      "[删除]: 删除【" + good.Name + "】，删前库存【" + strconv.FormatInt(good.Quantity, 10) + "】。",
		Status:    1,
	}
	return AddHistory(&his)
}

// 写入恢复货品历史记录
// 输入: 恢复货品快照, 备注
// 返回: id和err
func AddRecoverHistory(good *Good, remark string) (id int64, err error) {
	his := History{
		Event:     EVENT_RECOVER,
		GoodId:    good.Id,
		GoodName:  good.Name,
		GoodDesc:  good.Desc,
		GoodPrice: good.Price,
		GoodImage: good.Image,
		Quantity:  good.Quantity,
		Money:     float64(good.Quantity) * good.Price,
		Remark:    remark,
		Info:      "[恢复]: 恢复【" + good.Name + "】，库存【" + strconv.FormatInt(good.Quantity, 10) + "】。",
		Status:    1,
	}
	return AddHistory(&his)
}

// 写入更新货品历史记录
// 输入: 更新货品快照, 旧货品快照, 备注
// 返回: id和err
func AddUpdateHistory(good *Good, oldGood *Good, remark string) (id int64, err error) {
	updateMap := make(map[string]string)
	if good.Name != oldGood.Name{
		updateMap["名称"] = "\"" + oldGood.Name + "\" 更改为 \"" + good.Name + "\""
	}
	if good.Desc != oldGood.Desc{
		updateMap["描述"] = "\"" + oldGood.Desc + "\" 更改为 \"" + good.Desc + "\""
	}
	if good.Price != oldGood.Price{
		updateMap["价格"] = "\"" + strconv.FormatFloat(oldGood.Price, 'f', 2, 64) + "\" 更改为 \"" +
			strconv.FormatFloat(good.Price, 'f', 2, 64) + "\""
	}
	if good.Quantity != oldGood.Quantity{
		updateMap["数量"] = "\"" + strconv.FormatInt(oldGood.Quantity, 10) + "\" 更改为 \"" +
			strconv.FormatInt(good.Quantity, 10) + "\""
	}
	if good.Image != oldGood.Image{
		updateMap["图片"] = "\"" + oldGood.Image + "\" 更改为 \"" + good.Image + "\""
	}
	if good.Deleted != oldGood.Deleted {
		oldGoodDeleteString := "正常"
		goodDeleteString := "已恢复"
		if oldGood.Deleted == true{
			oldGoodDeleteString = "已删除"
		}
		if good.Deleted == true{
			goodDeleteString = "已删除"
		}
		updateMap["删除状态"] = "\"" + oldGoodDeleteString + "\" 更改为 \"" + goodDeleteString + "\""
	}
	infoString := "[更新]: "
	c := 0
	length := len(updateMap)
	for key, value := range updateMap{
		c++
		infoString += key + ": 由" + value
		if c < length{
			infoString += ", "
		}
	}
	his := History{
		Event:     EVENT_UPDATE,
		GoodId:    good.Id,
		GoodName:  good.Name,
		GoodDesc:  good.Desc,
		GoodPrice: good.Price,
		GoodImage: good.Image,
		Quantity:  good.Quantity,
		Money:     float64(good.Quantity) * good.Price,
		Remark:    remark,
		Info:      infoString,
		Status:    1,
	}
	return AddHistory(&his)
}

// 读取售货历史表
// 输入: 开始时间, 结束时间, 事件, 名称, 排序(ORDER)
// 返回: history序列, err
func GoodHistoryByName(startTime time.Time, endTime time.Time, event string, name string, order int) (mh []History, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	queryCond := ormHandle.QueryTable(&History{})

	if !startTime.IsZero() {
		queryCond = queryCond.Filter("create_time__gte", startTime)
	}
	if !endTime.IsZero() {
		queryCond = queryCond.Filter("create_time__lte", endTime)
	}
	if name != "" {
		queryCond = queryCond.Filter("good_name__icontains", name)
	}
	if event != "" {
		queryCond = queryCond.Filter("event__contains", event)
	}

	if order == ORDERBY_UNKNOWN || order == ORDERBY_DESC {
		queryCond = queryCond.OrderBy("-create_time")
	} else if order == ORDERBY_ASC {
		queryCond = queryCond.OrderBy("create_time")
	} else {
		return nil, errors.New(ORDERBY_ARGUMENT_ERROR)
	}

	if _, err = queryCond.All(&mh); err != nil {
		return nil, err
	}
	return mh, nil
}

// 使用货品ID取得售货历史
// 输入: 开始时间, 结束时间, 货品ID, 顺序
// 输出: history列表, err
func GoodHistory(startTime *time.Time, endTime *time.Time, id int64, order int) (result []History, err error){
	o := orm.NewOrmUsingDB(DBNAME)
	qc := o.QueryTable(&History{})

	qc = qc.Filter("create_time__gte", startTime).
		Filter("create_time__lte", endTime)
	if id > 0 {
		qc = qc.Filter("good_id", id)
	}

	if order == ORDERBY_UNKNOWN || order == ORDERBY_DESC {
		qc = qc.OrderBy("-create_time")
	}else if order == ORDERBY_ASC {
		qc = qc.OrderBy("create_time")
	}else{
		return nil, errors.New(ORDERBY_ARGUMENT_ERROR)
	}

	if _, err = qc.All(&result); err != nil{
		return nil, err
	}
	return result, err
}

func StatEventSummary(startTime *time.Time, endTime *time.Time, name string, event string, stat int32)(result map[string]float64, err error){
	o := orm.NewOrmUsingDB(DBNAME)

	sql := "Select {SELECT} from history where {WHERE}"
	sel := make([]string, 0, 4)
	whe := make([]string, 0, 5)

	if event != "" {
		whe = append(whe, "event = '"+strings.Replace(event, "'", "", -1)+"'")
	}
	if name = strings.Replace(name, "'", "", -1); name != "" {
		whe = append(whe, "name like '%" + name + "%'")
	}
	whe = append(whe, "create_time >= '" + startTime.Format(common.WiredTime) + "'")
	whe = append(whe, "create_time <= '" + endTime.Format(common.WiredTime) + "'")
	whe = append(whe, "status = 1")

	if stat & STAT_SUM_MONEY > 0 {
		sel = append(sel, "sum(money) as " + STAT_SUM_MONEY_KEY)
	}
	if stat & STAT_SUM_QUANTITY > 0 {
		sel = append(sel, "sum(quantity) as " + STAT_SUM_QUANTITY_KEY)
	}
	if stat & STAT_SUM_PROFITS > 0 {
		sel = append(sel, "sum(money - good_price) as " + STAT_SUM_PROFITS_KEY)
	}
	if stat & STAT_COUNT_ITEMS > 0 {
		sel = append(sel, "count(1) as " + STAT_COUNT_ITEMS_KEY)
	}

	if len(sel) <= 0{
		return nil, errors.New("no stat items set")
	}

	sql = strings.Replace(sql, "{SELECT}", strings.Join(sel, ","), 1)
	sql = strings.Replace(sql, "{WHERE}", strings.Join(whe, " and "), 1)

	resultInterface := make([]orm.Params, 0)
	_, err = o.Raw(sql).Values(&resultInterface)

	if err != nil{
		return
	}

	result = make(map[string]float64)

	if len(resultInterface) <= 0{
		return
	}

	for key, value := range resultInterface[0]{
		if value == nil {
			result[key] = 0
		}else {
			result[key], err = strconv.ParseFloat(value.(string), 64)
			if err != nil {
				result[key] = 0
			}
		}
	}
	return
}

func StatSoldSummary(startTime *time.Time, endTime *time.Time, name string, stat int32)(result map[string]float64, err error){
	return StatEventSummary(startTime, endTime, name, EVENT_SELL, stat)
}

func StatImportedSummary(startTime *time.Time, endTime *time.Time, name string, stat int32)(result map[string]float64, err error){
	return StatEventSummary(startTime, endTime, name, EVENT_IMPORT, stat)
}

func StatExportedSummary(startTime *time.Time, endTime *time.Time, name string, stat int32)(result map[string]float64, err error){
	return StatEventSummary(startTime, endTime, name, EVENT_EXPORT, stat)
}

func StatDeletedSummary(startTime *time.Time, endTime *time.Time, name string, stat int32)(result map[string]float64, err error){
	return StatEventSummary(startTime, endTime, name, EVENT_DELETE, stat)
}

// 售货历史通过Event获取
// 使用下面4个函数获得相应的售货历史事件的实际值
// 输入: 开始时间, 结束时间, 物品名称(模糊搜索), event, 想要获取哪些stat
// 输出: map[id] map[summingitem] value
func StatEvent(startTime *time.Time, endTime *time.Time, name string, event string, stat int32) (result map[int64]map[string]float64, err error) {
	o := orm.NewOrmUsingDB(DBNAME)

	sql := "Select {SELECT} from history where {WHERE} group by {GROUP_BY}"
	sel := make([]string, 0, 4)
	whe := make([]string, 0, 5)
	grpby := make([]string, 0, 1)

	whe = append(whe, "event = '" + strings.Replace(event, "'", "", -1) + "'")
	whe = append(whe, "create_time >= '" + startTime.Format(common.WiredTime) + "'")
	whe = append(whe, "create_time <= '" + endTime.Format(common.WiredTime) + "'")
	if name = strings.Replace(name, "'", "\\'", -1); name != "" {
		whe = append(whe, "name like '%" + name + "%'")
	}
	whe = append(whe, "status = 1")

	sel = append(sel, "good_id as id")
	grpby = append(grpby, "good_id")
	if stat & STAT_SUM_MONEY > 0 {
		sel = append(sel, "sum(money) as " + STAT_SUM_MONEY_KEY)
	}
	if stat & STAT_SUM_QUANTITY > 0 {
		sel = append(sel, "sum(quantity) as " + STAT_SUM_QUANTITY_KEY)
	}
	if stat & STAT_SUM_PROFITS > 0 {
		sel = append(sel, "sum(money - good_price) as " + STAT_SUM_PROFITS_KEY)
	}
	if stat & STAT_COUNT_ITEMS > 0 {
		sel = append(sel, "count(1) as " + STAT_COUNT_ITEMS_KEY)
	}

	if len(sel) <= 0{
		return nil, errors.New("no stat items set")
	}

	sql = strings.Replace(sql, "{SELECT}", strings.Join(sel, ","), 1)
	sql = strings.Replace(sql, "{WHERE}", strings.Join(whe, " and "), 1)
	sql = strings.Replace(sql, "{GROUP_BY}", strings.Join(grpby, ","), 1)

	resultInterface := make([]orm.Params, 0)
	_, err = o.Raw(sql).Values(&resultInterface)

	if err != nil{
		return
	}

	result = make(map[int64]map[string]float64)

	if len(resultInterface) <= 0 {
		//return nil, errors.New("database return nil statistic results")
		return
	}

	for i, r := range resultInterface{
		// result[i] = make(map[string]float64)
		goodId, ok := r["id"]
		if !ok {
			fmt.Println("models.StatEvent SQL result row[" + strconv.Itoa(i) + "] not has key 'id', ignored.")
			continue
		}
		goodIdString, ok := goodId.(string)
		if !ok {
			fmt.Println("models.StatEvent SQL result row[" + strconv.Itoa(i) + "] key 'id' cannot parse to int: '" + goodId.(string) + "', ignored.")
			continue
		}
		goodIdInt, transErr := strconv.ParseInt(goodIdString, 10, 64)
		if transErr != nil {
			fmt.Println("models.StatEvent SQL result row[" + strconv.Itoa(i) + "] key 'id' cannot parse to int: '" + goodIdString + "', ignored.")
			continue
		}
		for key, value := range r{
			if _, ok := result[goodIdInt]; !ok {
				result[goodIdInt] = make(map[string]float64)
			}
			result[goodIdInt][key], err = strconv.ParseFloat(value.(string), 64)
			if err != nil{
				result[goodIdInt][key] = 0
			}
		}
	}
	return
}

// 售卖历史统计
func StatSoldGoods(startTime *time.Time, endTime *time.Time, name string, stat int32) (result map[int64]map[string]float64, err error) {
	return StatEvent(startTime, endTime, name, EVENT_SELL, stat)
}

// 进货统计
func StatImportedGoods(startTime *time.Time, endTime *time.Time, name string, stat int32) (result map[int64]map[string]float64, err error) {
	return StatEvent(startTime, endTime, name, EVENT_IMPORT, stat)
}

// 撤柜统计
func StatExportedGoods(startTime *time.Time, endTime *time.Time, name string, stat int32) (result map[int64]map[string]float64, err error) {
	return StatEvent(startTime, endTime, name, EVENT_EXPORT, stat)
}

// 删除统计
func StatDeletedGoods(startTime *time.Time, endTime *time.Time, name string, stat int32) (result map[int64]map[string]float64, err error) {
	return StatEvent(startTime, endTime, name, EVENT_DELETE, stat)
}

//恢复统计
func StatRecoveredGoods(startTime *time.Time, endTime *time.Time, name string, stat int32) (result map[int64]map[string]float64, err error){
	return StatEvent(startTime, endTime, name, EVENT_RECOVER, stat)
}

func StatSellGoodsByGoodId(startTime *time.Time, endTime *time.Time, ids []int64, stat int32) (result map[int64]map[string]float64, err error){
	return StatEventByGoodIds(startTime, endTime, ids, EVENT_SELL, stat)
}

func StatEventByGoodIds(startTime *time.Time, endTime *time.Time, ids []int64, event string, stat int32) (result map[int64]map[string]float64, err error){
	o := orm.NewOrmUsingDB(DBNAME)

	sql := "Select {SELECT} from history where {WHERE} group by {GROUP_BY}"
	sel := make([]string, 0, 4)
	whe := make([]string, 0, 5)
	grpby := make([]string, 0, 1)

	whe = append(whe, "event = '" + strings.Replace(event, "'", "", -1) + "'")
	whe = append(whe, "create_time >= '" + startTime.Format(common.WiredTime) + "'")
	whe = append(whe, "create_time <= '" + endTime.Format(common.WiredTime) + "'")
	if ids != nil && len(ids) > 0{
		idStr := make([]string, 0, len(ids))
		for _, id := range ids{
			idStr = append(idStr, strconv.FormatInt(id, 10))
		}
		whe = append(whe, "good_id in (" + strings.Join(idStr, ",") + ")")
	}
	whe = append(whe, "status = 1")

	sel = append(sel, "good_id as id")
	grpby = append(grpby, "good_id")
	if stat & STAT_SUM_MONEY > 0 {
		sel = append(sel, "sum(money) as " + STAT_SUM_MONEY_KEY)
	}
	if stat & STAT_SUM_QUANTITY > 0 {
		sel = append(sel, "sum(quantity) as " + STAT_SUM_QUANTITY_KEY)
	}
	if stat & STAT_SUM_PROFITS > 0 {
		sel = append(sel, "sum(money - good_price) as " + STAT_SUM_PROFITS_KEY)
	}
	if stat & STAT_COUNT_ITEMS > 0 {
		sel = append(sel, "count(1) as " + STAT_COUNT_ITEMS_KEY)
	}

	if len(sel) <= 0{
		return nil, errors.New("no stat items set")
	}

	sql = strings.Replace(sql, "{SELECT}", strings.Join(sel, ","), 1)
	sql = strings.Replace(sql, "{WHERE}", strings.Join(whe, " and "), 1)
	sql = strings.Replace(sql, "{GROUP_BY}", strings.Join(grpby, ","), 1)

	resultInterface := make([]orm.Params, 0)
	_, err = o.Raw(sql).Values(&resultInterface)

	if err != nil{
		return
	}

	result = make(map[int64]map[string]float64)

	if len(resultInterface) <= 0 {
		//return nil, errors.New("database return nil statistic results")
		return
	}

	for i, r := range resultInterface{
		// result[i] = make(map[string]float64)
		goodId, ok := r["id"]
		if !ok {
			fmt.Println("models.StatEvent SQL result row[" + strconv.Itoa(i) + "] not has key 'id', ignored.")
			continue
		}
		goodIdString, ok := goodId.(string)
		if !ok {
			fmt.Println("models.StatEvent SQL result row[" + strconv.Itoa(i) + "] key 'id' cannot parse to string: '" + goodId.(string) + "', ignored.")
			continue
		}
		goodIdInt, transErr := strconv.ParseInt(goodIdString, 10, 64)
		if transErr != nil {
			fmt.Println("models.StatEvent SQL result row[" + strconv.Itoa(i) + "] key 'id' cannot parse to int: '" + goodIdString + "', ignored.")
			continue
		}

		for key, value := range r{
			if _, ok := result[goodIdInt]; !ok {
				result[goodIdInt] = make(map[string]float64)
			}
			result[goodIdInt][key], err = strconv.ParseFloat(value.(string), 64)
			if err != nil{
				result[goodIdInt][key] = 0
			}
		}
	}
	return
}
