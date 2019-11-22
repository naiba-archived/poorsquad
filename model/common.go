package model

import "time"

// Common ..
type Common struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
