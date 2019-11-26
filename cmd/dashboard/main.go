package main

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/patrickmn/go-cache"

	"github.com/naiba/poorsquad/controller"
	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
	"github.com/naiba/poorsquad/service/github"
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
	if cf.Debug {
		db = db.Debug()
	}

	db.AutoMigrate(model.User{}, model.Company{}, model.UserCompany{},
		model.Account{}, model.Team{}, model.Repository{}, model.UserRepository{},
		model.UserTeam{}, model.TeamRepository{}, model.AccountRepository{})

	dao.Init(db, cache.New(5*time.Minute, 10*time.Minute), cf)

	go controller.RunWeb()
	go github.SyncAll(db)
	select {}
}
