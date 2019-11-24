package model

import "time"

const (
	_ = iota
	// UTPOutsideMember 小组外部成员
	UTPOutsideMember
	// UTPMember 小组成员
	UTPMember
	// UTPManager 小组管理员
	UTPManager
)

// UserTeam ..
type UserTeam struct {
	UserID     uint64 `gorm:"primary_key;auto_increment:false"`
	TeamID     uint64 `gorm:"primary_key;auto_increment:false"`
	Permission uint64
	UpdatedAt  time.Time
}
