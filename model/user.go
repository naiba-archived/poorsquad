package model

import "time"

// User ...
type User struct {
	Common    `json:"common,omitempty"`
	Login     string `json:"login,omitempty"`      // 登录名
	AvatarURL string `json:"avatar_url,omitempty"` // 头像地址
	Name      string `json:"name,omitempty"`       // 昵称
	Blog      string `json:"blog,omitempty"`       // 网站链接
	Email     string `json:"email,omitempty"`      // 邮箱
	Hireable  bool   `json:"hireable,omitempty"`
	Bio       string `json:"bio,omitempty"` // 个人简介

	Token        string    // 认证 Token
	TokenExpired time.Time // Token 过期时间
	SuperAdmin   *bool     `json:"super_admin,omitempty"` // 超级管理员
}
