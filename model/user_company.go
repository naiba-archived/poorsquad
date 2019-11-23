package model

import "time"

const (
	_ = iota
	// UCPMember 企业成员
	UCPMember
	// UCPManager 企业管理员
	UCPManager
	// UCPSuperManager 企业超级管理员
	UCPSuperManager
)

// UserCompany ..
type UserCompany struct {
	UserID     uint64 `gorm:"primary_key;auto_increment:false"`
	CompanyID  uint64 `gorm:"primary_key;auto_increment:false"`
	Permission uint64
	CreatedAt  time.Time
}
