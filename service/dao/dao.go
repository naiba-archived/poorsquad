package dao

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/naiba/avalanche"
	"github.com/patrickmn/go-cache"

	"github.com/naiba/poorsquad/model"
)

// DB ..
var DB *gorm.DB

// Conf ..
var Conf *model.Config

// Cache ..
var Cache *cache.Cache

// Init ..
func Init(dbp *gorm.DB, cp *cache.Cache, cf *model.Config) {
	DB = dbp
	Cache = cp
	Conf = cf
}

// GetUserByID ..
func GetUserByID(id uint64) (user model.User, err error) {
	key := fmt.Sprintf("user%d", id)
	var userI interface{}
	var ok bool
	userI, ok = Cache.Get(key)
	if !ok {
		userI, err = avalanche.Do(key, func() (interface{}, error) {
			var u model.User
			e := DB.Where("id = ?", id).First(&u).Error
			return u, e
		})
	}
	if err == nil {
		user = userI.(model.User)
	}
	return
}
