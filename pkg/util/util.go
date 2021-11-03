package util

import (
	"fmt"
	"net"
	"os"
)

func InterFaceIP() (string,error) {
	addr := "eth0"
	ipq := os.Getenv("OS_IP")
	if ipq == "" {
		return addr, fmt.Errorf("没有ip的env")
	}
	interfaces , err := net.Interfaces()
	if err != nil {
		return addr, fmt.Errorf("获取本地网卡error，%v", err)
	}

	for _, i := range interfaces {
		ips, _ := i.Addrs()
		for _, ip := range ips {
			if ipq == ip.String()[:len(ip.String())-3] {
				addr = i.Name
				break
			}
		}
	}

	return addr, nil
}
