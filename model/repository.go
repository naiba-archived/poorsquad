package model

import (
	"strconv"
	"time"

	"github.com/google/go-github/v28/github"
	"github.com/jinzhu/gorm"
)

// Repository ..
type Repository struct {
	ID          uint64 `gorm:"primary_key"`
	Name        string
	HTMLURL     string
	Description string
	Private     bool

	SyncedAt  time.Time // 最后一次同步
	AccountID uint64

	OutsideCollaborators []User
}

// NewRepositoryFromGitHub ..
func NewRepositoryFromGitHub(gr *github.Repository) Repository {
	var r Repository
	r.ID = uint64(gr.GetID())
	r.Name = gr.GetName()
	r.Private = gr.GetPrivate()
	r.HTMLURL = gr.GetHTMLURL()
	r.Description = gr.GetDescription()
	return r
}

// SID ..
func (r *Repository) SID() string {
	return strconv.FormatUint(r.ID, 10)
}

// SAccountID ..
func (r *Repository) SAccountID() string {
	return strconv.FormatUint(r.AccountID, 10)
}

type userIDres struct {
	UserID uint64
}

// RelatedOutsideCollaborators ..
func (r *Repository) RelatedOutsideCollaborators(db *gorm.DB) error {
	var ids []userIDres
	err := db.Raw(`SELECT user_repositories.user_id
	FROM user_repositories
	WHERE user_repositories.user_id
		NOT IN (
			SELECT user_teams.user_id FROM user_teams
			WHERE user_teams.team_id
			IN (
				SELECT team_repositories.team_id FROM team_repositories
				WHERE team_repositories.repository_id = ?
			)
		)
	AND user_repositories.repository_id = ?`, r.ID, r.ID).Scan(&ids).Error
	if err != nil {
		return err
	}
	for i := 0; i < len(ids); i++ {
		var u User
		u.ID = ids[i].UserID
		r.OutsideCollaborators = append(r.OutsideCollaborators, u)
	}
	return nil
}

// IsOutsideCollaborator 是不是只在一个 Team 的开发者
func (r *Repository) IsOutsideCollaborator(db *gorm.DB, userID uint64) (bool, error) {
	if len(r.OutsideCollaborators) == 0 {
		if err := r.RelatedOutsideCollaborators(db); err != nil {
			return false, err
		}
	}
	for i := 0; i < len(r.OutsideCollaborators); i++ {
		if r.OutsideCollaborators[i].ID == userID {
			return true, nil
		}
	}
	return false, nil
}

// IsIndividualCollaborator 是不是只在一个 Team 的开发者
func (r *Repository) IsIndividualCollaborator(db *gorm.DB, user *User) (bool, error) {
	teams, err := r.GetTeams(db)
	if err != nil {
		return false, err
	}
	var count int
	for i := 0; i < len(teams); i++ {
		for j := 0; j < len(user.TeamsID); j++ {
			if teams[i] == user.TeamsID[j] {
				count++
			}
		}
	}
	return count < 2, nil
}

// GetTeams ..
func (r *Repository) GetTeams(db *gorm.DB) ([]uint64, error) {
	// 取拥有仓库的 team 列表
	var teamRepositories []TeamRepository
	if err := db.Where("repository_id = ?", r.ID).Find(&teamRepositories).Error; err != nil {
		return nil, err
	}
	var teamIDs []uint64
	for i := 0; i < len(teamRepositories); i++ {
		teamIDs = append(teamIDs, teamRepositories[i].TeamID)
	}
	return teamIDs, nil
}
