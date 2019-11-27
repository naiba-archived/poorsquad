package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// Team ..
type Team struct {
	Common
	Name string `json:"name,omitempty"`

	CompanyID uint64 `json:"company_id,omitempty"`

	Repositories []uint64 `gorm:"-"`
}

// CheckUserPermission ..
func (t *Team) CheckUserPermission(db *gorm.DB, userID uint64, permission uint64) (uint64, error) {
	var ut UserTeam
	if err := db.Where("team_id = ? AND user_id = ? AND permission > ?", t.ID, userID, permission).First(&ut).Error; err != nil {
		return 0, fmt.Errorf("您不是该小组的雇员(%s)", err)
	}
	if permission > 0 && ut.Permission < permission {
		return 0, fmt.Errorf("您不是该小组的管理人员(%d)", ut.Permission)
	}
	return ut.Permission, nil
}

// FetchRepositories ..
func (t *Team) FetchRepositories(db *gorm.DB) {
	var repos []TeamRepository
	db.Where("team_id = ?", t.ID).Find(&repos)
	for i := 0; i < len(repos); i++ {
		t.Repositories = append(t.Repositories, repos[i].RepositoryID)
	}
}
