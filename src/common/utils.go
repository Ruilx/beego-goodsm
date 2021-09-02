package common

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"mime/multipart"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-basic/uuid"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"

	"github.com/jackpal/gateway"
	"github.com/skip2/go-qrcode"
)

const WiredTime = "2006-01-02 15:04:05"
const WiredDate = "2006-01-02"

const (
	IMAGE_UPLOAD_PATH        = "static/upload/"
	IMAGE_ORIGIN_PATH_PREFIX = IMAGE_UPLOAD_PATH + "origin/"
	IMAGE_THUMB_PATH_PREFIX  = IMAGE_UPLOAD_PATH + "thumb/"
)

const MINE_CONTENT_TYPE = "Content-Type"
const ACCEPTED_MIME_CONTENT_TYPE = "image/bmp;image/png;image/jpeg;image/jpg"

const (
	THUMB_IMG_WIDTH  = 267
	THUMB_IMG_HEIGHT = 150
)

const SAVING_IMAGE_SUFFIX = ".jpg"

func QRCodeImageBase64(msg string) (imageBase64 string, err error) {
	pngImage, err := qrcode.Encode(msg, qrcode.Medium, 150)
	if err != nil {
		return
	}
	pngBase64 := base64.StdEncoding.EncodeToString(pngImage)
	return "data:image/png;base64," + pngBase64, nil
}

func GetIPAddresses() (ip []*net.IPNet, err error) {
	var addrs []net.Addr
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return nil, err
	}
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = append(ip, ipnet)
			}
		}
	}
	return ip, nil
}

func GetIPAddresses2() (ip []string, err error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, i := range interfaces {
		if i.Flags&net.FlagUp != 0 && (i.Flags&net.FlagLoopback) == 0 {
			addrs, err := i.Addrs()
			if err != nil {
				continue
			}
			for _, j := range addrs {
				if ipnet, ok := j.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil {
						ip = append(ip, ipnet.IP.String())
					}
				}
			}
		}
	}
	return
}

func GetActiveIPAddress() (ip net.IP, err error) {
	return gateway.DiscoverInterface()
}

func GetActiveIPGateway() (ip net.IP, err error) {
	return gateway.DiscoverGateway()
}

func RerenderImage(fileHandler multipart.File, fileHandlerHeader *multipart.FileHeader) (filename string, err error) {
	if fileHandler == nil {
		return "", errors.New("file handler invalid")
	}
	if _, err = fileHandler.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	var wdDir string
	if wdDir, err = os.Getwd(); err != nil {
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
	if !isAcceptMimeContentType {
		return "", errors.New("unsupported mime-type ContentType: " + mimeContentType)
	}
	var img image.Image
	var imgThumb image.Image
	var outImg *os.File
	var outThumb *os.File
	if mimeContentType == "image/jpg" || mimeContentType == "image/jpeg" {
		img, err = jpeg.Decode(fileHandler)
	} else if mimeContentType == "image/png" {
		img, err = png.Decode(fileHandler)
	} else if mimeContentType == "image/bmp" {
		img, err = bmp.Decode(fileHandler)
	} else {
		err = errors.New("unsupported mime-type ContentType: " + mimeContentType)
	}
	if err != nil {
		return "", err
	}
	size := img.Bounds().Size()
	if size.X/size.Y >= THUMB_IMG_WIDTH/THUMB_IMG_HEIGHT {
		imgThumb = resize.Resize(uint(THUMB_IMG_WIDTH), uint(size.Y)*uint(THUMB_IMG_WIDTH)/uint(size.X), img, resize.Lanczos3)
	} else {
		imgThumb = resize.Resize(uint(size.X*THUMB_IMG_HEIGHT/size.Y), uint(THUMB_IMG_HEIGHT), img, resize.Lanczos3)
	}
	var uuidStr string
	if uuidStr, err = uuid.GenerateUUID(); err != nil {
		return "", errors.New("Create UUID failed: " + err.Error())
	}
	filename = uuidStr + SAVING_IMAGE_SUFFIX
	originFilename := wdDir + IMAGE_ORIGIN_PATH_PREFIX + uuidStr + SAVING_IMAGE_SUFFIX
	if outImg, err = os.Create(originFilename); err != nil {
		return "", errors.New("create original image file failed: " + err.Error())
	}
	defer outImg.Close()
	thumbFilename := wdDir + IMAGE_THUMB_PATH_PREFIX + uuidStr + SAVING_IMAGE_SUFFIX
	if outThumb, err = os.Create(thumbFilename); err != nil {
		return "", errors.New("create thumb image file failed: " + err.Error())
	}
	defer outThumb.Close()
	if err = jpeg.Encode(outImg, img, nil); err != nil {
		return "", errors.New("writing original file failed: " + err.Error())
	}
	if err = jpeg.Encode(outThumb, imgThumb, nil); err != nil {
		return "", errors.New("writing thumb file failed: " + err.Error())
	}

	_, err = outImg.Seek(0, io.SeekStart)
	if err != nil {
		return filename, errors.New("outImg file seek to 0 failed: " + err.Error())
	}
	md5Obj := md5.New()
	buf := make([]byte, 1024)
	n := 1
	for n > 0 {
		if n, err = outImg.Read(buf); err != nil && err != io.EOF {
			return filename, errors.New("outImg file read err. err: " + err.Error())
		}
		if n <= 0 {
			break
		}
		if _, err = md5Obj.Write(buf); err != nil {
			return filename, errors.New("outImg md5 checksum append error. err: " + err.Error())
		}
	}
	md5Str := hex.EncodeToString(md5Obj.Sum(nil))

	newOriginFilename := wdDir + IMAGE_ORIGIN_PATH_PREFIX + md5Str + SAVING_IMAGE_SUFFIX
	_ = outImg.Close()
	if err = os.Rename(originFilename, newOriginFilename); err != nil {
		return filename, errors.New("outImg cannot rename to md5sum. err: " + err.Error())
	}
	newThumbFilename := wdDir + IMAGE_THUMB_PATH_PREFIX + md5Str + SAVING_IMAGE_SUFFIX
	_ = outThumb.Close()
	if err = os.Rename(thumbFilename, newThumbFilename); err != nil {
		return filename, errors.New("outThumb cannot rename to md5sum, err: " + err.Error())
	}

	return md5Str + SAVING_IMAGE_SUFFIX, nil
}

func NumberUnitFormat(number int64, prec int8, unit int, baseUnit int, glue string) (result string, err error) {
	sizeTable := []string{"", "K", "M", "G", "T", "P", "E", "Z", "Y", "B", "N", "D"}
	count := baseUnit
	unitf := float64(unit)
	numberf := float64(number)
	err = nil
	for numberf > unitf || -numberf >= unitf {
		count += 1
		numberf /= unitf
	}
	if count >= len(sizeTable) {
		return "", errors.New("number is too big to show and calculate")
	}
	if math.Floor(numberf) == numberf {
		result = strconv.FormatInt(int64(numberf), 10) + glue + sizeTable[count]
	} else {
		result = fmt.Sprintf("%.*f", prec, numberf) + glue + sizeTable[count]
	}
	return
}

func Struct2Map(stru interface{}, lowerKey bool) (mp map[string]interface{}, err error) {
	val := reflect.ValueOf(stru)
	typ := reflect.TypeOf(stru)
	fieldNum := val.NumField()
	fieldNum2 := typ.NumField()
	mp = make(map[string]interface{})
	if fieldNum != fieldNum2 {
		return nil, errors.New("same struct has not same field size")
	}
	for i := 0; i < fieldNum; i++ {
		name := typ.Field(i).Name
		if lowerKey {
			name = strings.ToLower(name)
		}
		valu := val.Field(i)
		switch valu.Type().Name() {
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
		case "bool":
			mp[name] = valu.Bool()
		default:
			mp[name] = valu.String()
		}
	}
	return mp, nil
}

func IpMatch(ips []*net.IPNet, gateway *net.IP) (matchedIps []net.IP) {
	for _, ip := range ips {
		ipStu := ip.IP
		ipMask := ip.Mask
		ipGatewayNet := &net.IPNet{IP: *gateway, Mask: ipMask}
		if ipGatewayNet.Contains(ipStu) {
			matchedIps = append(matchedIps, ipStu)
		}
	}
	return
}

func IsMobileUsingUserAgent(ua string) (isMobile bool) {
	uaPartList := strings.Split(ua, " ")
	isMobile = true
	final := false
	for _, uaPart := range uaPartList{
		if strings.Contains(uaPart, "iPhone"){
			isMobile = true
			final = true
		}
		if strings.Contains(uaPart, "Android"){
			isMobile = true
			final = true
		}
		if strings.Contains(uaPart, "Mobile"){
			isMobile = true
			final = true
		}
		if strings.Contains(uaPart, "Windows"){
			isMobile = false
			final = true
		}
		if strings.Contains(uaPart, "Macintosh"){
			isMobile = false
			final = true
		}
		if final{
			return
		}
	}
	return
}

func GetTimeBeginning(key string, t1 *time.Time)(t time.Time, err error){
	switch key{
	case "second":
		t = time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), t1.Minute(), t1.Second(), 0, time.Local)
	case "minute":
		t = time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), t1.Minute(), 0, 0, time.Local)
	case "hour":
		t = time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), 0, 0, 0, time.Local)
	case "day":
		t = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.Local)
	case "month":
		t = time.Date(t1.Year(), t1.Month(), 1, 0, 0, 0, 0, time.Local)
	case "year":
		t = time.Date(t1.Year(), 1, 1, 0, 0, 0, 0, time.Local)
	default:
		return *t1, errors.New("key is not valid. choices:(second, minute, hour, day, month, year)")
	}
	return t, nil
}