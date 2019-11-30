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

// FetchCompanyManagers ..
func FetchCompanyManagers(company *model.Company) {
	var userCompanies []model.UserCompany
	DB.Where("company_id = ?", company.ID).Find(&userCompanies)
	for i := 0; i < len(userCompanies); i++ {
		user, _ := GetUserByID(userCompanies[i].UserID)
		if userCompanies[i].Permission == model.UCPManager {
			company.Managers = append(company.Managers, user)
		} else {
			company.SuperManagers = append(company.SuperManagers, user)
		}
	}
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
			Cache.SetDefault(key, u)
			return u, e
		})
	}
	if err == nil {
		user = userI.(model.User)
	}
	return
}
