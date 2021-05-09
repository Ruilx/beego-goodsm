package main

import (
	"fmt"
	"net"
)

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

func main() {
	var ip []string
	var err error
	if ip, err = GetIPAddress(); err != nil {
		fmt.Println(err.Error())
	}
	for _, v := range ip {
		fmt.Println(v)
	}
}
