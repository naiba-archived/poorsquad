package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/naiba/poorsquad/controller"
	"github.com/naiba/poorsquad/model"
)

func main() {
	cf, err := model.ReadInConfig("data/config.yaml")
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open("sqlite3", "data/github.db")
	if err != nil {
		panic(err)
	}
	db = db.Debug()
	db.AutoMigrate(model.User{})
	go controller.RunWeb(cf, db)
	select {}
}
