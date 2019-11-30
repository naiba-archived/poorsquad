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

	RepositoriesID []uint64 `gorm:"-"`
	Repositories   []Repository
	EmployeesID    []uint64 `gorm:"-"`
	Employees      []User
	Managers       []User
}

// CheckUserPermission ..
func (t *Team) CheckUserPermission(db *gorm.DB, userID uint64, minPermission uint64) (uint64, error) {
	var ut UserTeam
	if err := db.Where("team_id = ? AND user_id = ?", t.ID, userID).First(&ut).Error; err != nil {
		return 0, fmt.Errorf("您不是该小组的雇员(%s)", err)
	}
	if minPermission > 0 && ut.Permission < minPermission {
		return 0, fmt.Errorf("权限不足(%d < %d)", ut.Permission, minPermission)
	}
	return ut.Permission, nil
}

// FetchRepositoriesID ..
func (t *Team) FetchRepositoriesID(db *gorm.DB) {
	var repos []TeamRepository
	db.Where("team_id = ?", t.ID).Find(&repos)
	for i := 0; i < len(repos); i++ {
		t.RepositoriesID = append(t.RepositoriesID, repos[i].RepositoryID)
	}
}

// FetchRepositories ..
func (t *Team) FetchRepositories(db *gorm.DB) {
	db.Where("id IN (?)", t.RepositoriesID).Find(&t.Repositories)
}

// FetchEmployeesID ..
func (t *Team) FetchEmployeesID(db *gorm.DB) {
	var uts []UserTeam
	db.Where("team_id = ?", t.ID).Find(&uts)
	for i := 0; i < len(uts); i++ {
		t.EmployeesID = append(t.EmployeesID, uts[i].UserID)
	}
}

// FetchEmployees ..
func (t *Team) FetchEmployees(db *gorm.DB) {
	db.Where("id IN (?)", t.EmployeesID).Find(&t.Employees)
}
