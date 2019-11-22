package main

import (
	"github.com/naiba/poorsquad/controller"
	"github.com/naiba/poorsquad/model"
)

func main() {
	cf, err := model.ReadInConfig("data/config.yaml")
	if err != nil {
		panic(err)
	}
	go controller.RunWeb(cf)
	select {}
}
