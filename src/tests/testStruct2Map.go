package main

import (
	"errors"
	"fmt"
	"reflect"
)

// (mp map[string]interface{})

func Struct2Map(stru interface{})(mp map[string]interface{}, err error){
	val := reflect.ValueOf(stru)
	typ := reflect.TypeOf(stru)
	fieldNum := val.NumField()
	fieldNum2 := typ.NumField()
	mp = make(map[string]interface{})
	if fieldNum != fieldNum2{
		return nil, errors.New("same struct has not same field size")
	}
	for i := 0; i < fieldNum; i++{
		name := typ.Field(i).Name
		valu := val.Field(i)
		switch valu.Type().Name(){
		case "int", "int8", "int16", "int32", "int64", "byte", "rune":
			mp[name] = valu.Int()
		case "uint", "uint8", "uint16", "uint32", "uint64":
			mp[name] = valu.Uint()
		case "float32", "float64":
			mp[name] = valu.Float()
		case "complex64", "complex128":
			mp[name] = valu.Complex()
		case "string":
			mp[name] = valu.String()
		}
	}
	return mp, nil
}

type A struct{
	a int
	b int64
	c uint
	d uint64
	e float32
	f float64
	g string
	h *A
	i chan int
	j map[string]string
	k interface{}
	l complex128
}

func main(){
	a := A{
		a: 1,
		b: 2,
		c: 3,
		d: 4,
		e: 5.0,
		f: 6.0,
		g: "STRING",
		j: make(map[string]string),
		l: complex(123,123),
	}
	mp, err := Struct2Map(a)
	if err != nil{
		panic(err)
	}
	fmt.Println(mp)
}