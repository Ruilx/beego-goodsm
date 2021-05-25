package common

import (
	"errors"
	"github.com/go-basic/uuid"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net"
	"os"
	"strings"
)

const (
	IMAGE_UPLOAD_PATH = "/static/upload/"
	IMAGE_ORIGIN_PATH_PREFIX = IMAGE_UPLOAD_PATH + "origin/"
	IMAGE_THUMB_PATH_PREFIX = IMAGE_UPLOAD_PATH + "thumb/"
)

const MINE_CONTENT_TYPE = "Content-Type"
const ACCEPTED_MIME_CONTENT_TYPE = "image/bmp;image/png;image/jpeg;image/jpg"

const (
	THUMB_IMG_WIDTH = 267
	THUMB_IMG_HEIGHT = 150
)

const SAVING_IMAGE_SUFFIX = ".jpg"



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
	var wdDir string
	if wdDir, err = os.Getwd(); err != nil{
		return "", err
	}

	mimeContentType := fileHandlerHeader.Header.Get(MINE_CONTENT_TYPE)
	acceptedMimeContentType := strings.Split(ACCEPTED_MIME_CONTENT_TYPE, ";")
	isAcceptMimeContentType := false
	for _, i := range acceptedMimeContentType {
		if mimeContentType == i {
			isAcceptMimeContentType = true
			break
		}
	}
	if !isAcceptMimeContentType{
		return "", errors.New("unsupported mime-type ContentType: " + mimeContentType)
	}
	var img image.Image
	var imgThumb image.Image
	var outImg *os.File
	var outThumb *os.File
	if mimeContentType == "image/jpg" || mimeContentType == "image/jpeg" {
		img, err = jpeg.Decode(fileHandler)
	}else if mimeContentType == "image/png"{
		img, err = png.Decode(fileHandler)
	}else if mimeContentType == "image/bmp" {
		img, err = bmp.Decode(fileHandler)
	}else{
		err = errors.New("unsupported mime-type ContentType: " + mimeContentType)
	}
	if err != nil {
		return "", err
	}
	size := img.Bounds().Size()
	if size.X / size.Y >= THUMB_IMG_WIDTH / THUMB_IMG_HEIGHT {
		imgThumb = resize.Resize(uint(THUMB_IMG_WIDTH), uint(size.Y) * uint(THUMB_IMG_WIDTH) / uint(size.X), img, resize.Lanczos3)
	}else{
		imgThumb = resize.Resize(uint(size.X * THUMB_IMG_HEIGHT / size.Y), uint(THUMB_IMG_HEIGHT), img, resize.Lanczos3)
	}
	var uuidStr string
	if uuidStr, err = uuid.GenerateUUID(); err != nil{
		return "", errors.New("Create UUID failed: " + err.Error())
	}
	if outImg, err = os.Create(wdDir + IMAGE_ORIGIN_PATH_PREFIX + uuidStr + SAVING_IMAGE_SUFFIX); err != nil{
		return "", errors.New("create original image file failed: " + err.Error())
	}
	defer outImg.Close()
	if outThumb, err = os.Create(wdDir + IMAGE_THUMB_PATH_PREFIX + uuidStr + SAVING_IMAGE_SUFFIX); err != nil{
		return "", errors.New("create thumb image file failed: " + err.Error())
	}
	defer outThumb.Close()
	if err = jpeg.Encode(outImg, img, nil); err != nil{
		return "", errors.New("writing original file failed: " + err.Error())
	}
	if err = jpeg.Encode(outThumb, imgThumb, nil); err != nil{
		return "", errors.New("writing thumb file failed: " + err.Error())
	}
	return uuidStr + SAVING_IMAGE_SUFFIX, nil
}