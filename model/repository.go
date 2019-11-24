package model

import (
	"strconv"
	"time"

	"github.com/google/go-github/v28/github"
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
