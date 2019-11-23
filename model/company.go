package model

// Company ...
type Company struct {
	Common    `json:"common,omitempty"`
	Brand     string `json:"brand,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`

	ProjectCount  uint64 `json:"project_count,omitempty"`
	EmployeeCount uint64 `json:"employee_count,omitempty"`
}
