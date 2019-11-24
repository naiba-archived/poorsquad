package model

// Company ...
type Company struct {
	Common
	Brand     string `json:"brand,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`

	ProjectCount  uint64 `json:"project_count,omitempty"`
	EmployeeCount uint64 `json:"employee_count,omitempty"`
	TeamCount     uint64 `json:"team_count,omitempty"`
}
