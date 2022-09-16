package main

import (
	"flag"
	"fmt"
	"github.com/fangjie-luoxi/git-hook/config"
	"net/http"

	"github.com/fangjie-luoxi/git-hook/log"
	"github.com/fangjie-luoxi/git-hook/parse"
)

//go:generate go build -ldflags="-w -s"

func main() {
	cfg := flag.String("c", "./config.yaml", "配置文件地址")
	flag.Parse()
	fmt.Println("配置文件地址:", *cfg)
	err := config.Setup(*cfg)
	if err != nil {
		fmt.Println("读取配置文件失败, err:", err.Error())
	}

	log.NewPrivateLog(log.Config{
		Console:         true,
		StorageLocation: "./logs/",
		Level:           "debug",
	})
	http.HandleFunc("/", parse.Handle)
	port := "11000"
	if config.Config.Port != "" {
		port = config.Config.Port
	}
	log.Info("服务启动: 0.0.0.0:", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Errorf("http server failed, err:%v\n", err)
		return
	}
	log.Info("服务关闭")
}
