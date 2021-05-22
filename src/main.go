package main

import (
	"beego-goodsm/common"
	_ "beego-goodsm/routers"

	beego "github.com/beego/beego/v2/server/web"

	"fmt"
	"strconv"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	port, err := beego.AppConfig.Int("httpport")
	if err != nil {
		fmt.Println("Expect an httpport in config, if not set it will set to 8080.")
		port = 8080
	}
	ips, err := common.GetIPAddress()
	if err != nil {
		fmt.Println("Cannot get any IPs from your computer, try 'localhost' in your browser.")
	}
	fmt.Println("Connect IPs: ==========")
	for _, v := range ips {
		fmt.Println(v + ":" + strconv.Itoa(port))
	}
	fmt.Println("=======================")

	beego.Run()
}
