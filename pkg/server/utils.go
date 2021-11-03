package server

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	mb    = 1024 * 1024 * 30
	total = 1000
	rxNet1 uint64
	txNet1 uint64
	rxNet2 uint64
	txNet2 uint64
)


func GetNet(device string) {
	var err error
	m := [3]int{0, 0, 0}
	for {
		rxNet1, err = commandR(device)
		if err != nil {
			logrus.Infoln("ifconfig command err : ",err.Error())
			continue
		}
		txNet1, err = commandT(device)
		if err != nil {
			logrus.Infoln("ifconfig command err : ",err.Error())
			continue
		}

		time.Sleep(30 * time.Second)

		rxNet2, err = commandR(device)
		if err != nil {
			logrus.Infoln("ifconfig command err : ",err.Error())
			continue
		}
		txNet2, err = commandT(device)
		if err != nil {
			logrus.Infoln("ifconfig command err : ",err.Error())
			continue
		}

		m[2] = int(rxNet2-rxNet1) / mb + int(txNet2-txNet1) / mb

		if total -  int(0.1 * float64(m[0]) + 0.2 *float64(m[1]) + 0.7 * float64(m[2])) < 0 {
			TotalBytes = 0
		}else {
			TotalBytes = total -  int(0.1 * float64(m[0]) + 0.2 *float64(m[1]) + 0.7 * float64(m[2]))
		}
		m[0], m[1] = m[1], m[2]
	}
}

func commandT(device string) (uint64, error) {

	com := "ifconfig " + device + ` | awk '/TX packets/{print $5}'`
	cmd := exec.Command("/bin/bash", "-c", com)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return 0, err
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println("wait err :", err.Error())
	}
	number, err := strconv.ParseUint(strings.TrimSpace(string(bytes)), 10, 64)
	if err != nil {
		return 0, err
	}
	return number, nil
}

func commandR(device string) (uint64, error) {

	com := "ifconfig " + device + ` | awk '/RX packets/{print $5}'`
	cmd := exec.Command("/bin/bash", "-c", com)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return 0, err
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println("wait err :", err.Error())
	}
	number, err := strconv.ParseUint(strings.TrimSpace(string(bytes)), 10, 64)
	if err != nil {
		return 0, err
	}
	return number, nil

}



//func getMaxBand() int {
//	//TODO 获取网卡信息 根据网卡类型设置total值
//	return 0
//}
