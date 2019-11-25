package github

import (
	"context"
	"time"

	GitHubAPI "github.com/google/go-github/v28/github"
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"

	"github.com/naiba/poorsquad/model"
)

var (
	db *gorm.DB
)

// SyncAll ..
func SyncAll(d *gorm.DB) {
	if d != nil {
		db = d
	}
	if db == nil {
		panic("nil db instance")
	}
	var accounts []model.Account
	db.Find(&accounts)
	for i := 0; i < len(accounts); i++ {
		account := accounts[i]
		if account.SyncedAt.Add(time.Minute * 30).After(time.Now()) {
			continue
		}
		account.SyncedAt = time.Now()
		db.Save(&account)
		go Sync(db, &account, account.Token)
	}
}

// NewAPIClient ..
func NewAPIClient(ctx context.Context, token string) *GitHubAPI.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	return GitHubAPI.NewClient(tc)
}

// Sync ..
func Sync(db *gorm.DB, account *model.Account, token string) {
	var errInfo string
	defer func() {
		// 最终步骤，更新当前账号的最后同步时间
		data := model.Account{
			Status:   model.ASOK,
			SyncedAt: time.Now(),
		}
		if errInfo != "" {
			data.Status = model.ASFail
			data.Message = errInfo
		}
		db.Model(account).Updates(data)
	}()
	ctx := context.Background()
	client := NewAPIClient(ctx, token)
	// ========================= Repositories 同步 =========================
	// 1. 取数据库中仓库列表
	var reposInDB []model.Repository
	db.Where("account_id = ?", account.ID).Find(&reposInDB)
	var repos []*GitHubAPI.Repository
	next := 1
	for next != 0 {
		opt := &GitHubAPI.RepositoryListOptions{
			Type: "owner",
		}
		opt.PerPage = 100
		opt.Page = next

		reposInner, resp, err := client.Repositories.List(ctx, "", opt)

		next = resp.NextPage
		if err != nil {
			errInfo = err.Error()
			return
		}
		repos = append(repos, reposInner...)
	}
	// 2. 查找要删除的 Repo 进行清理
	for j := 0; j < len(reposInDB); j++ {
		for i := 0; i < len(repos); i++ {
			if uint64(repos[i].GetID()) == reposInDB[j].ID {
				continue
			}
		}
		db.Delete(reposInDB[j])
		reposInDB = append(reposInDB[:j], reposInDB[j+1:]...)
	}
	// 3. 查找要追加的 Repo 进行追加
	for i := 0; i < len(repos); i++ {
		for j := 0; j < len(reposInDB); j++ {
			if uint64(repos[i].GetID()) == reposInDB[j].ID {
				continue
			}
		}
		repo := model.NewRepositoryFromGitHub(repos[i])
		repo.AccountID = account.ID
		reposInDB = append(reposInDB, repo)
	}
	// ========================= Collaborators 同步 =========================
	for i := 0; i < len(reposInDB); i++ {
		repo := reposInDB[i]
		var userRepos []model.UserRepository
		db.Where("repository_id = ?", repo.ID).Find(&userRepos)
		var cos []*GitHubAPI.User
		// 1. 拉取 Collaborators
		nextCr := 1
		for nextCr != 0 {
			optCr := &GitHubAPI.ListCollaboratorsOptions{}
			optCr.PerPage = 100
			optCr.Page = next
			cosInner, respCr, err := client.Repositories.ListCollaborators(ctx, account.Login, repo.Name, optCr)
			nextCr = respCr.NextPage
			if err != nil {
				errInfo = err.Error()
				return
			}
			cos = append(cos, cosInner...)
		}
		// 2. 用户入库
		for j := 0; j < len(cos); j++ {
			newUser := model.NewUserFromGitHub(cos[j])
			var oldUser model.User
			if err := db.Where("id = ?", newUser.ID).First(&oldUser).Error; err == nil {
				if oldUser.Token == "" {
					newUser.IssueNewToken()
				} else {
					newUser.Token = oldUser.Token
				}
			} else {
				newUser.IssueNewToken()
			}
			db.Save(&newUser)
		}
		// 3. 查找要删除的 Collaborators 进行清理
		for j := 0; j < len(userRepos); j++ {
			for k := 0; k < len(cos); k++ {
				if userRepos[j].UserID == uint64(cos[k].GetID()) {
					continue
				}
			}
			db.Delete(userRepos[j])
			userRepos = append(userRepos[:j], userRepos[j+1:]...)
		}
		// 4. 查找要追加的 Collaborators
		for k := 0; k < len(cos); k++ {
			if uint64(cos[k].GetID()) == account.ID {
				// 越过用户本身
				continue
			}
			for j := 0; j < len(userRepos); j++ {
				if userRepos[j].UserID == uint64(cos[k].GetID()) {
					continue
				}
			}
			var ur model.UserRepository
			ur.UserID = uint64(cos[k].GetID())
			ur.RepositoryID = repo.ID
			ur.AccountID = account.ID
			userRepos = append(userRepos, ur)
		}
		// 5. Collaborators 入库
		for k := 0; k < len(userRepos); k++ {
			db.Save(userRepos[k])
		}
		// 6. Repository 入库
		repo.SyncedAt = time.Now()
		db.Save(repo)
	}
}
