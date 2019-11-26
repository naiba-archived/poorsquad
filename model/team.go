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
}

// CheckUserPermission ..
func (t *Team) CheckUserPermission(db *gorm.DB, userID uint64, permission uint64) (uint64, error) {
	var ut UserTeam
	if err := db.Where("tem_id = ? AND user_id = ? AND permission > ?", t.ID, userID, permission).First(&ut).Error; err != nil {
		return 0, fmt.Errorf("您不是该小组的雇员(%s)", err)
	}
	if permission > 0 && ut.Permission < permission {
		return 0, fmt.Errorf("您不是该小组的管理人员(%d)", ut.Permission)
	}
	return ut.Permission, nil
}
