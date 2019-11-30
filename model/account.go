package model

import (
	"time"

	"github.com/google/go-github/v28/github"
)

const (
	_ = iota
	// ASOK ..
	ASOK
	// ASFail ..
	ASFail
)

// Account ..
type Account struct {
	Common    `json:"common,omitempty"`
	Login     string `gorm:"UNIQUE_INDEX" json:"login,omitempty"`
	Name      string `json:"name,omitempty"` // 昵称
	AvatarURL string `json:"avatar_url,omitempty"`

	Status  uint   `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`

	SyncedAt  time.Time
	CompanyID uint64 `json:"company_id,omitempty"`
}

// NewAccountFromGitHub ..
func NewAccountFromGitHub(gu *github.User) Account {
	var u Account
	u.ID = uint64(gu.GetID())
	u.Login = gu.GetLogin()
	u.AvatarURL = gu.GetAvatarURL()
	u.Name = gu.GetName()
	u.Status = ASOK
	// 昵称为空的情况
	if u.Name == "" {
		u.Name = u.Login
	}
	return u
}
