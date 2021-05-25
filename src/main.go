package main

import (
	"beego-goodsm/common"
	_ "beego-goodsm/routers"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/pkg/errors"

	"fmt"
	"os"
	"strconv"
)

func createDirectory(path string) (err error){
	var info os.FileInfo
	beelogs := logs.GetBeeLogger()
	if info, err = os.Stat(path); err != nil{
		if !os.IsNotExist(err){
			return
		}
	}else{
		if !info.IsDir(){
			if err = os.Remove(path); err != nil{
				return errors.New("failed while trying to remove not expected file to create a directory, path: " + path)
			}
			beelogs.Warn("[Init] Removed a file which has a same name to a directory when try to create the directory called: " + path)
		}else{
			return nil
		}
	}
	if _, err = os.Stat(path); err != nil && !os.IsNotExist(err){
		return
	}
	if err = os.Mkdir(path, os.ModeDir); err != nil{
		return
	}
	if err = os.Chmod(path, os.ModeDir); err != nil {
		return
	}
	beelogs.Info("[Init] Created directory: " + path)
	return nil
}

func initialize() (err error){
	var dir string
	if dir, err = os.Getwd(); err != nil{
		return
	}
	if dir == ""{
		return errors.New("failed while trying to find current directory. executable path: " + dir)
	}
	if err = createDirectory(dir + common.IMAGE_UPLOAD_PATH); err != nil{
		return
	}
	if err = createDirectory(dir + common.IMAGE_ORIGIN_PATH_PREFIX); err != nil{
		return
	}
	if err = createDirectory(dir + common.IMAGE_THUMB_PATH_PREFIX); err != nil{
		return
	}
	return nil
}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/static"] = "static"
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

	if err = initialize(); err != nil{
		fmt.Println("Cannot initialing webserver environment with error: '" + err.Error() + "', please contract your administrator for further information.")
		os.Exit(1)
	}

	beego.Run()
}
