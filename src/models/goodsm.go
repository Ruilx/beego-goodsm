package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

const DBNAME = "default"

//type Good struct {
//	Id int32 `orm:"column:id;type:int;AUTO_INCREMENT;not null"`
//	Name string `orm:"column:name;type:varchar(32);index:name_i"`
//	Desc string `orm:"column:desc;type:text"`
//	Price float64 `orm:"column:price;type:double;default (0)"`
//	Quantity int64 `orm:"column:quantity;type:BIGINT;not null"`
//	CreateTime time.Time `orm:"column:create_time;type:TIMESTAMP;default (datetime('now', 'localtime'))"`
//}

/** CREATE TABLE SQL:
CREATE TABLE "main"."NewTable" (
"id"  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
"name"  varchar(32) NOT NULL,
"desc"  TEXT,
"price"  REAL,
"quantity"  INTEGER,
"create_at"  datetime,
"update_at"  datetime,
"delete_at"  datetime
);
CREATE INDEX "main"."name_i"
ON "goods" ("name" ASC);
 */

const (
	ORDERBY_UNKNOWN = 0
	ORDERBY_ASC
	ORDERBY_DESC
)

type Good struct {
	Id int32
	Name string `orm:"size(32)"`
	Desc string
	Price float64
	Quantity int64
}

func init(){
	var (
		dbPath string
		err error
	)

	if dbPath, err = web.AppConfig.String("db_path"); err != nil{
		panic("Cannot find 'dbPath' in app config: " + err.Error())
	}
	fmt.Println("DBPATH: " + dbPath)

	if err = orm.RegisterDriver("sqlite3", orm.DRSqlite); err != nil {
		panic("Cannot register sqlite3 for model: " + err.Error())
	}

	if err = orm.RegisterDataBase(DBNAME, "sqlite3", dbPath); err != nil{
		panic("Cannot register database " + DBNAME + " for sqlite3: " + err.Error())
	}
	orm.RegisterModel(new(Good))
	//if err = orm.RunSyncdb(DBNAME, false, true); err != nil {
	//	panic("Run sync db failed, error: " + err.Error())
	//}
}


func AddGoods(goods *Good)(id int64, err error){
	fmt.Println("Good: ", goods)
	ormHandle := orm.NewOrmUsingDB(DBNAME)
	id, err = ormHandle.Insert(goods)
	return
}

func GetGoodsById(id int32)(res *Good, err error){
	ormHandle := orm.NewOrm()
	result := &Good{Id: id}
	if err = ormHandle.QueryTable(new(Good)).Filter("Id", id).RelatedSel().One(result); err != nil{
		return nil, err
	}
	return result, err
}

func GetGoods(query map[string]string, fields []string, sortBy []string, order []int, offset int64, limit int64)(ml []interface{}, err error){
	ormHandle := orm.NewOrm()
	queryCond := ormHandle.QueryTable(new(Good))

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

func UpdateGoodsById(good *Good)(id int64, err error){
	ormHandle := orm.NewOrm()
	idReady := Good{Id: good.Id}
	if err = ormHandle.Read(&idReady); err != nil {
		if id, err = ormHandle.Update(good); err == nil {
			return
		}
	}
	return
}

func DeleteGoodsById(id int32)(ok bool, err error){
	ormHandle := orm.NewOrm()
	good := Good{Id: id}
	if err = ormHandle.Read(&good); err == nil{
		return false, err
	}
	return true, err
}
