package model

// Team ..
type Team struct {
	Common
	Name string `json:"name,omitempty"`

	CompanyID uint64 `json:"company_id,omitempty"`

	Employees        []User
	OutsideEmployees []User
}
