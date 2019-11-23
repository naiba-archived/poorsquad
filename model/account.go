package model

// Account ..
type Account struct {
	Common
	Login     string
	Name      string `json:"name,omitempty"` // 昵称
	Token     string
	AvatarURL string
	Status    uint

	CompanyID uint64
	UserID    uint64
}
