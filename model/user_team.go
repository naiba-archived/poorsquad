package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

const (
	_ = iota
	// UTPOutsideMember 小组外部成员
	UTPOutsideMember
	// UTPMember 小组成员
	UTPMember
	// UTPManager 小组管理员
	UTPManager
)

// UserTeam ..
type UserTeam struct {
	UserID     uint64 `gorm:"primary_key;auto_increment:false"`
	TeamID     uint64 `gorm:"primary_key;auto_increment:false"`
	Permission uint64
	UpdatedAt  time.Time
}

// GetIndividualFromTeams ..
func GetIndividualFromTeams(db *gorm.DB, teams []uint64) ([]uint64, error) {
	// 取独立用户列表
	type Result struct {
		UserID uint64
		Num    uint64
	}
	var res []Result
	if err := db.Raw("SELECT user_teams.user_id, COUNT(user_teams.user_id) AS num FROM user_teams WHERE team_id in (?)", teams).Scan(&res).Error; err != nil {
		return nil, err
	}
	var individual []uint64
	for i := 0; i < len(res); i++ {
		if res[i].Num == 1 {
			individual = append(individual, res[i].UserID)
		}
	}
	return individual, nil
}
