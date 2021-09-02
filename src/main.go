package main

import (
	"beego-goodsm/common"
	_ "beego-goodsm/routers"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/pkg/errors"
)

const (
	CONF_QRCODE_IP      = "qrcode_ip"
	CONF_QRCODE_PNG_B64 = "qrcode_base64"
	CONF_HTTP_PORT      = "httpport"
)

func checkIpQrcode(ipstr string) {
	ipInConf, err := beego.AppConfig.String(CONF_QRCODE_IP)
	if err != nil {
		ipInConf = ""
	}
	if ipstr != ipInConf {
		qrImgBase64, err := common.QRCodeImageBase64(ipstr)
		if err != nil {
			fmt.Println("Cannot draw a qrcode to save and use in webpage. err: ", err.Error())
			return
		}
		if err = beego.AppConfig.Set(CONF_QRCODE_PNG_B64, qrImgBase64); err != nil {
			fmt.Println("Cannot save qrcode base64 in config file. err: ", err.Error())
		}else{
			fmt.Println("Saved config: " + CONF_QRCODE_PNG_B64)
		}
		if err = beego.AppConfig.Set(CONF_QRCODE_IP, ipstr); err != nil {
			fmt.Println("Cannot save qrcode ip in config file. err: ", err.Error())
		}else{
			fmt.Println("Saved config: " + CONF_QRCODE_IP)
		}
	} else {
		fmt.Println("QRCode '" + ipstr + "' has already in config.")
	}
}

func createDirectory(path string) (err error) {
	var info os.FileInfo
	beelogs := logs.GetBeeLogger()
	if info, err = os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !info.IsDir() {
			if err = os.Remove(path); err != nil {
				return errors.New("failed while trying to remove not expected file to create a directory, path: " + path)
			}
			beelogs.Warn("[Init] Removed a file which has a same name to a directory when try to create the directory called: " + path)
		} else {
			return nil
		}
	}
	if _, err = os.Stat(path); err != nil && !os.IsNotExist(err) {
		return
	}
	if err = os.Mkdir(path, os.ModeDir); err != nil {
		return
	}
	if err = os.Chmod(path, os.ModeDir); err != nil {
		return
	}
	beelogs.Info("[Init] Created directory: " + path)
	return nil
}

func initialize() (err error) {
	var dir string
	if dir, err = os.Getwd(); err != nil {
		return
	}
	if dir == "" {
		return errors.New("failed while trying to find current directory. executable path: " + dir)
	}
	if err = createDirectory(dir + "/" + common.IMAGE_UPLOAD_PATH); err != nil {
		return
	}
	if err = createDirectory(dir + "/" + common.IMAGE_ORIGIN_PATH_PREFIX); err != nil {
		return
	}
	if err = createDirectory(dir + "/" + common.IMAGE_THUMB_PATH_PREFIX); err != nil {
		return
	}
	beego.BConfig.WebConfig.AutoRender = true
	beego.BConfig.EnableGzip = true

	return nil
}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/static"] = "static"
	}
	if beego.AppConfig == nil {
		fmt.Println("Config lost, program will exit.")
		os.Exit(1)
	}
	port, err := beego.AppConfig.Int(CONF_HTTP_PORT)
	if err != nil {
		fmt.Println("Expect an httpport in config, if not set it will set to 8080.")
		port = 8080
	}
	ips, err := common.GetIPAddresses()
	var ipsStr []string
	if err != nil {
		fmt.Println("Cannot get any IPs from your computer, try 'localhost' or '127.0.0.1' in your browser.")
		ips := make([]net.IP, 1)
		ipBytes := make(net.IP, net.IPv4len)
		copy(ipBytes, []byte{127, 0, 0, 1})
		ips = append(ips, ipBytes)
	} else {
		for _, ip := range ips {
			ipsStr = append(ipsStr, ip.IP.String())
		}
	}
	ip4, err := common.GetActiveIPAddress()
	if err != nil {
		fmt.Println("GetActiveIPAddress Failed. Error: " + err.Error())
		fmt.Println("Cannot get IP from your computer, try 'localhost' in your browser.")
		// ip4 = []byte{127, 0, 0, 1}
		if ip, err := common.GetActiveIPGateway(); err == nil {
			if ips4 := common.IpMatch(ips, &ip); len(ips4) > 0 {
				ip4 = ips4[0]
			} else {
				ip4 = []byte{127, 0, 0, 1}
			}
		} else {
			fmt.Println(err)
			ip4 = []byte{127, 0, 0, 1}
		}

	}
	ip := ip4.String()
	fmt.Println("Connect IPs: ==========")
	for _, v := range ipsStr {
		fmt.Print(v + ":" + strconv.Itoa(port))
		if v == ip {
			fmt.Println(" <-- [Active]")
		} else {
			fmt.Println()
		}
	}
	fmt.Println("=======================")

	if ip, err := common.GetActiveIPGateway(); err == nil {
		fmt.Println("Gatway IP: ", ip.String())
	}

	fmt.Println("Checking active IP to config and try to draw qrcode.")
	serverUrl := "http://" + ip + ":" + strconv.Itoa(port)
	checkIpQrcode(serverUrl)

	if err = initialize(); err != nil {
		fmt.Println("Cannot initialing webserver environment with error: '" + err.Error() + "', please contract your administrator for further information.")
		os.Exit(1)
	}

	if ok, err := beego.AppConfig.Bool("open_browser"); err == nil && ok {
		fmt.Println("Ready to open browser: ok:", ok, "err:", err)
		err = common.OpenUrlUsingCommand(serverUrl)
		if err != nil {
			fmt.Println("Open browser failed: " + err.Error())
		}
	}

	fmt.Println("Starting Beego server...")
	beego.Run()
}
