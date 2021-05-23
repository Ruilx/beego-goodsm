package common

import (
	"errors"
	"image"
	"io"
	"net"
	"mime/multipart"
	"os"
)

const IMAGE_CREATE_PATH_PREFIX = "/static/upload/"

func GetIPAddress() (ip []string, err error) {
	var addrs []net.Addr
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return nil, err
	}
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = append(ip, ipnet.IP.String())
			}
		}
	}
	return ip, nil
}

func RerenderImage(fileHandler multipart.File, fileHandlerHeader *multipart.FileHeader) (filename string, err error) {
	if fileHandler == nil{
		return "", errors.New("file handler invalid")
	}
	if _, err = fileHandler.Seek(0, io.SeekStart); err != nil{
		return "", err
	}
	filename = IMAGE_CREATE_PATH_PREFIX + "123.png"
	var file *os.File
	if file, err = os.Create(filename); err != nil{
		return "", err
	}
	defer file.Close()
	var img image.Image
	fileHandlerHeader.Header



}