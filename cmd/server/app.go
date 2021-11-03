package main

import (
	"9nml-device-plugin/pkg/server"
	"9nml-device-plugin/pkg/util"
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
	"os"
	"path/filepath"
	"time"
)

func main() {

	deviceName := flag.String("i", "eth0", "default net device")
	flag.Parse()
	interFace, err := util.InterFaceIP()
	if err != nil {
		interFace = *deviceName
	}
	go server.GetNet(interFace)
	log.Info("easyalgo device plugin starting")
	easyalgoSrv := server.NewEasyalgoServer()
	go easyalgoSrv.Run()

	// 向 kubelet 注册
	if err := easyalgoSrv.RegisterToKubelet(); err != nil {
		log.Fatalf("register to kubelet error: %v", err)
	} else {
		log.Infoln("register to kubelet successfully")
	}

	// 监听 kubelet.sock，一旦创建则重启
	devicePluginSocket := filepath.Join(server.DevicePluginPath, server.KubeletSocket)
	log.Info("device plugin socket name:", devicePluginSocket)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Failed to created FS watcher.")
		os.Exit(1)
	}
	defer watcher.Close()
	err = watcher.Add(server.DevicePluginPath)
	if err != nil {
		log.Error("watch kubelet error")
		return
	}
	log.Info("watching kubelet.sock")
	for {
		select {
		case event := <-watcher.Events:
			log.Infof("watch kubelet events: %s, event name: %s, isCreate: %v", event.Op.String(), event.Name, event.Op&fsnotify.Create == fsnotify.Create)
			if event.Name == devicePluginSocket && event.Op&fsnotify.Create == fsnotify.Create {
				time.Sleep(time.Second)
				log.Fatalf("inotify: %s created, restarting.", devicePluginSocket)
			}
		case err := <-watcher.Errors:
			log.Fatalf("inotify: %s", err)
		}
	}
}
