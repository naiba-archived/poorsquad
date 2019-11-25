package model

import "time"

// TeamRepository ..
type TeamRepository struct {
	TeamID       uint64 `gorm:"primary_key;auto_increment:false"`
	RepositoryID uint64 `gorm:"primary_key;auto_increment:false"`
	UpdatedAt    time.Time
}
