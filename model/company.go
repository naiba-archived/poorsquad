package model

// Company ...
type Company struct {
	Common
	Brand     string `json:"brand,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`

	UserID uint64 `gorm:"index" json:"user_id,omitempty"`
}
