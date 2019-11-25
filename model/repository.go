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

	OutsideUsers []User
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

// RelatedOutsideUsers ..
func (r *Repository) RelatedOutsideUsers(db *gorm.DB) {
	var ids []userIDres
	db.Raw(`SELECT user_repositories.user_id
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
	AND user_repositories.repository_id = ?`, r.ID, r.ID).Scan(&ids)
	for i := 0; i < len(ids); i++ {
		var u User
		u.ID = ids[i].UserID
		r.OutsideUsers = append(r.OutsideUsers, u)
	}
}
