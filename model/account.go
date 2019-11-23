package model

import "github.com/google/go-github/v28/github"

// Account ..
type Account struct {
	Common
	Login     string `json:"login,omitempty"`
	Name      string `json:"name,omitempty"` // 昵称
	AvatarURL string `json:"avatar_url,omitempty"`

	Status uint   `json:"status,omitempty"`
	Token  string `json:"token,omitempty"`

	CompanyID uint64 `json:"company_id,omitempty"`
}

// NewAccountFromGitHub ..
func NewAccountFromGitHub(gu *github.User) Account {
	var u Account
	u.ID = uint64(gu.GetID())
	u.Login = gu.GetLogin()
	u.AvatarURL = gu.GetAvatarURL()
	u.Name = gu.GetName()
	// 昵称为空的情况
	if u.Name == "" {
		u.Name = u.Login
	}
	return u
}
