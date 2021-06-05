package models

import (
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
	ORDERBY_UNKNOWN = 0
	ORDERBY_ASC
	ORDERBY_DESC
)

const (
	STATUS_UNKNWON = 0 // 不使用
	STATUS_ACTIVE      // 默认, 状态有效
	STATUS_DELETED     // 状态被人工删除, 但未还原
	STATUS_REVOKE      // 状态被撤回(比如撤回卖出等)
)

const (
	EVENT_ADD    = "登记"
	EVENT_SELL   = "售出"
	EVENT_IMPORT = "进货"
	EVENT_EXPORT = "撤柜"
	EVENT_DELETE = "删除"
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

	orm.Debug = true

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
	// 注册表
	orm.RegisterModel(&Good{})
	orm.RegisterModel(&History{})
	if err = orm.RunSyncdb(DBNAME, false, true); err != nil {
		panic("Run sync db failed, error: " + err.Error())
	}
}

// 添加货物
func AddGoods(goods *Good) (id int64, err error) {
	fmt.Println("Good: ", goods)
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	id, err = ormHandle.Insert(goods)
	return
}

// 按照ID获取货物信息
func GetGoodsById(id int32) (res *Good, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	result := &Good{Id: id}
	if err = ormHandle.QueryTable(&Good{}).Filter("Id", id).RelatedSel().One(result); err != nil {
		return nil, err
	}
	return result, err
}

// 按照名称获取货物信息
func GetGoodsByName(name string) (res *Good, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	result := &Good{Name: name}
	if err = ormHandle.QueryTable(&Good{}).Filter("Name", name).RelatedSel().One(result); err != nil {
		return nil, err
	}
	return result, err
}

func GetGoods(name string, order int) (ml []Good, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	queryCond := ormHandle.QueryTable(&Good{})

	if name != "" {
		queryCond = queryCond.Filter("name__icontains", name)
	}
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

// 获得所有货物
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
func UpdateGoodsById(good *Good, col ...string) (id int64, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	idReady := Good{Id: good.Id}
	if err = ormHandle.Read(&idReady); err == nil {
		return ormHandle.Update(good, col ...)
	}
	return
}

// 按照ID删除货物
func DeleteGoodsById(id int32) (dbId int64, err error) {
	ormHandle := orm.NewOrm()
	good := Good{Id: id}
	if err = ormHandle.Read(&good); err == nil {
		return ormHandle.Delete(&good)
	}
	return
}

// 写入售货历史表
func AddHistory(his *History) (id int64, err error) {
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	id, err = ormHandle.Insert(his)
	return
}

func AddSellsHistory(good *Good, quantity int64, unitPrice float64, money float64, remark string, balance int64)(id int64, err error){
	his := History{
		Event: EVENT_SELL,
		GoodId: good.Id,
		GoodName: good.Name,
		GoodDesc: good.Desc,
		GoodPrice: good.Price,
		GoodImage: good.Image,
		Quantity: quantity,
		Money: money,
		Remark: remark,
		Info: "[出售]: 卖出【" + good.Name + "】，数量【" + strconv.FormatInt(quantity, 10) + "】，单价【" + strconv.FormatFloat(unitPrice, 'f', 2, 64) + "】，总价【" + strconv.FormatFloat(money, 'f', 2, 64,) + "】，剩余库存【" + strconv.FormatInt(balance, 10) + "】。",
		Status: 1,
	}
	return AddHistory(&his)
}

func AddImportHistory(good *Good, quantity int64, remark string, balance int64)(id int64, err error){
	money := good.Price * float64(quantity) // 进货物品所产生的定价价值
	his := History{
		Event: EVENT_IMPORT,
		GoodId: good.Id,
		GoodName: good.Name,
		GoodDesc: good.Desc,
		GoodPrice: good.Price,
		GoodImage: good.Image,
		Quantity: quantity,
		Money: money,
		Remark: remark,
		Info: "[进货]: 进货【" + good.Name + "】，数量【" + strconv.FormatInt(quantity, 10) + "】，库存【" + strconv.FormatInt(good.Quantity, 10) + "】，总价【" + strconv.FormatFloat(money, 'f', 2, 64) + "】，当前库存【" + strconv.FormatInt(balance, 10) + "】。",
		Status: 1,
	}
	return AddHistory(&his)
}

func AddExportHistory(good *Good, quantity int64, remark string, balance int64)(id int64, err error){
	money := good.Price * float64(quantity)
	his := History{
		Event: EVENT_EXPORT,
		GoodId: good.id,
		GoodName: good.Name,
		GoodDesc: good.Desc,
		GoodPrice: good.Price,
		GoodImage: good.Image,
		Quantity: quantity,
		Money: money,
		Remark: remark,
		Info: "[撤柜]: 撤柜【" + good.Name + "】，数量【" + strconv.FormatInt(quantity, 10) + "】，库存【" + strconv.FormatInt(good.Quantity, 10) + "】，总价【" + strconv.FormatFloat(money, 'f', 2, 64) + "】，当前库存【" + strconv.FormatInt(balance, 10) + "】。",
		Status: 1,
	}
	return AddHistory(&his)
}

func AddDeleteHistory(good *Good, remark string)(id int64, err error){
	his := History{
		Event: EVENT_DELETE,
		GoodId: Good.Id,
		GoodName: Good.Name,
		GoodDesc: Good.Desc,
		GoodPrice: Good.Price,
		GoodImage: Good.Image,
		Quantity: Good.Quantity,
		Money: float64(Good.Quantity) * Good.Price,
		Remark: remark,
		Info: "[删除]: 删除【" + good.Name + "】，删前库存【" + strconv.FormatInt(Good.Quantity, 10) + "】。"
		Status: 1,
	}
	return AddHistory(&his)
}

// 读取售货历史表
func getSellsHistory(startTime time.Time, endTime time.Time, name string, order int) (mh []History, err error) {
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

	if order == ORDERBY_UNKNOWN || order == ORDERBY_DESC {
		queryCond = queryCond.OrderBy("-create_time")
	} else if order == ORDERBY_ASC {
		queryCond = queryCond.OrderBy("create_time")
	} else {
		return nil, errors.New("Error for order, needs one of ORDERBY_ASC, ORDERBY_DESC, ORDERBY_UNKNOWN.")
	}

	if _, err = queryCond.All(&mh); err != nil {
		return nil, err
	}
	return mh, nil
}
