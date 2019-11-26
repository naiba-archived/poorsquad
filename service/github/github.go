package github

import (
	"context"
	"time"

	GitHubAPI "github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
)

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

// SyncAll ..
func SyncAll() {
	var accounts []model.Account
	dao.DB.Find(&accounts)
	for i := 0; i < len(accounts); i++ {
		account := accounts[i]
		if account.SyncedAt.Add(time.Minute * 30).After(time.Now()) {
			continue
		}
		account.SyncedAt = time.Now()
		dao.DB.Save(&account)
		ctx := context.Background()
		go SyncRepositories(ctx, NewAPIClient(ctx, account.Token), &account)
	}
}

// SyncRepositories ..
func SyncRepositories(ctx context.Context, client *GitHubAPI.Client, account *model.Account) {
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
		dao.DB.Model(account).Updates(data)
	}()

	// ========================= Repositories 同步 =========================
	// 1. 取数据库中仓库列表
	var reposInDB []model.Repository
	dao.DB.Where("account_id = ?", account.ID).Find(&reposInDB)
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
CHECKDEL:
	for j := 0; j < len(reposInDB); j++ {
		for i := 0; i < len(repos); i++ {
			if uint64(repos[i].GetID()) == reposInDB[j].ID {
				continue CHECKDEL
			}
		}
		dao.DB.Delete(reposInDB[j])
		reposInDB = append(reposInDB[:j], reposInDB[j+1:]...)
	}
	// 3. 查找要追加的 Repo 进行追加
CHECKADD:
	for i := 0; i < len(repos); i++ {
		for j := 0; j < len(reposInDB); j++ {
			if uint64(repos[i].GetID()) == reposInDB[j].ID {
				continue CHECKADD
			}
		}
		repo := model.NewRepositoryFromGitHub(repos[i])
		repo.AccountID = account.ID
		reposInDB = append(reposInDB, repo)
	}
	// ========================= Collaborators 同步 =========================
	for i := 0; i < len(reposInDB); i++ {
		repo := reposInDB[i]
		SyncCollaborator(ctx, client, account, &repo)
		// 6. Repository 入库
		repo.SyncedAt = time.Now()
		dao.DB.Save(repo)
	}
}

// SyncCollaborator ..
func SyncCollaborator(ctx context.Context, client *GitHubAPI.Client, account *model.Account, repo *model.Repository) error {
	var userRepos []model.UserRepository
	dao.DB.Where("repository_id = ?", repo.ID).Find(&userRepos)
	var cos []*GitHubAPI.User
	// 1. 拉取 Collaborators
	nextCr := 1
	for nextCr != 0 {
		optCr := &GitHubAPI.ListCollaboratorsOptions{}
		optCr.PerPage = 100
		optCr.Page = nextCr
		cosInner, respCr, err := client.Repositories.ListCollaborators(ctx, account.Login, repo.Name, optCr)
		nextCr = respCr.NextPage
		if err != nil {
			return err
		}
		cos = append(cos, cosInner...)
	}
	// 2. 雇员入库
	for j := 0; j < len(cos); j++ {
		newUser := model.NewUserFromGitHub(cos[j])
		var oldUser model.User
		if err := dao.DB.Where("id = ?", newUser.ID).First(&oldUser).Error; err == nil {
			if oldUser.Token == "" {
				newUser.IssueNewToken()
			} else {
				newUser.Token = oldUser.Token
			}
		} else {
			newUser.IssueNewToken()
		}
		dao.DB.Save(&newUser)
	}
	// 3. 查找要删除的 Collaborators 进行清理
CHECKDEL:
	for j := 0; j < len(userRepos); j++ {
		for k := 0; k < len(cos); k++ {
			if userRepos[j].UserID == uint64(cos[k].GetID()) {
				continue CHECKDEL
			}
		}
		dao.DB.Delete(userRepos[j])
		userRepos = append(userRepos[:j], userRepos[j+1:]...)
	}
	// 4. 查找要追加的 Collaborators
CHECKADD:
	for k := 0; k < len(cos); k++ {
		if uint64(cos[k].GetID()) == account.ID {
			// 越过雇员本身
			continue CHECKADD
		}
		for j := 0; j < len(userRepos); j++ {
			if userRepos[j].UserID == uint64(cos[k].GetID()) {
				continue
			}
		}
		var ur model.UserRepository
		ur.UserID = uint64(cos[k].GetID())
		ur.RepositoryID = repo.ID
		userRepos = append(userRepos, ur)
	}
	// 5. Collaborators 入库
	for k := 0; k < len(userRepos); k++ {
		dao.DB.Save(userRepos[k])
	}

	return nil
}

// TODO: 实现 Team 跟 Repo 的互通

// RemoveRepositoryFromTeam ..
func RemoveRepositoryFromTeam(team *model.Team, repositoryID uint64) []error {
	var errs []error
	return errs
}

// AddRepositoryFromTeam ..
func AddRepositoryFromTeam(team *model.Team, repositoryID uint64) []error {
	var errs []error
	return errs
}
