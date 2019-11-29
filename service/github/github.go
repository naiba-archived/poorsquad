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
		if account.SyncedAt.Add(time.Hour * 10).After(time.Now()) {
			continue
		}
		account.SyncedAt = time.Now()
		dao.DB.Save(&account)
		ctx := context.Background()
		go AccountSync(ctx, NewAPIClient(ctx, account.Token), &account)
	}
}

// AccountSync ..
func AccountSync(ctx context.Context, client *GitHubAPI.Client, account *model.Account) {
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
		if err != nil {
			errInfo = err.Error()
			return
		}
		next = resp.NextPage
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
		RepositorySync(ctx, client, account, &repo)
		// 6. Repository 入库
		repo.SyncedAt = time.Now()
		dao.DB.Save(repo)
	}
}

// RepositorySync ..
func RepositorySync(ctx context.Context, client *GitHubAPI.Client, account *model.Account, repo *model.Repository) error {
	var userRepos []model.UserRepository
	dao.DB.Where("repository_id = ?", repo.ID).Find(&userRepos)
	var cos []*GitHubAPI.User
	invitation := make(map[uint64]int64)
	// 1. 拉取 Collaborators
	nextCr := 1
	for nextCr != 0 {
		optCr := &GitHubAPI.ListCollaboratorsOptions{}
		optCr.PerPage = 100
		optCr.Page = nextCr
		cosInner, respCr, err := client.Repositories.ListCollaborators(ctx, account.Login, repo.Name, optCr)
		if err != nil {
			return err
		}
		nextCr = respCr.NextPage
		cos = append(cos, cosInner...)
	}
	var cov []*GitHubAPI.RepositoryInvitation
	// 1.2 拉取 Invitations
	nextCr = 1
	for nextCr != 0 {
		optCr := &GitHubAPI.ListOptions{}
		optCr.PerPage = 100
		optCr.Page = nextCr
		covInner, respCr, err := client.Repositories.ListInvitations(ctx, account.Login, repo.Name, optCr)
		if err != nil {
			return err
		}
		nextCr = respCr.NextPage
		cov = append(cov, covInner...)
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
	// 2.2 邀请中的人员入库
	for j := 0; j < len(cov); j++ {
		newUser := model.NewUserFromGitHub(cov[j].GetInvitee())
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
		invitation[newUser.ID] = cov[j].GetID()
		dao.DB.Save(&newUser)
		cos = append(cos, cov[j].GetInvitee())
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
		ur.InvitationID = invitation[ur.UserID]
		userRepos = append(userRepos, ur)
	}
	// 5. Collaborators 入库
	for k := 0; k < len(userRepos); k++ {
		dao.DB.Save(&userRepos[k])
	}
	return nil
}

// RemoveRepositoryFromTeam ..
func RemoveRepositoryFromTeam(ctx context.Context, client *GitHubAPI.Client, account *model.Account, team *model.Team, repository *model.Repository) error {
	teams, err := repository.GetTeams(dao.DB)
	if err != nil {
		return err
	}
	individual, err := model.GetIndividualFromTeams(dao.DB, teams)
	if err != nil {
		return err
	}
	var users []model.User
	if err = dao.DB.Where("id in (?)", individual).Find(&users).Error; err != nil {
		return err
	}
	// 从仓库中移除用户
	for i := 0; i < len(users); i++ {
		if err := RemoveEmployeeFromRepository(ctx, client, account, repository, &users[i]); err != nil {
			return err
		}
	}
	return nil
}

// AddRepositoryFromTeam ..
func AddRepositoryFromTeam(ctx context.Context, client *GitHubAPI.Client, account *model.Account, team *model.Team, repository *model.Repository) error {
	teams, err := repository.GetTeams(dao.DB)
	if err != nil {
		return err
	}
	individual, err := model.GetIndividualFromTeams(dao.DB, teams)
	if err != nil {
		return err
	}
	var users []model.User
	if err = dao.DB.Where("id in (?)", individual).Find(&users).Error; err != nil {
		return err
	}
	// 从仓库中移除用户
	for i := 0; i < len(users); i++ {
		if err := AddEmployeeToRepository(ctx, client, account, repository, &users[i]); err != nil {
			return err
		}
	}
	return nil
}

// AddEmployeeToTeam ..
func AddEmployeeToTeam(ctx context.Context, client *GitHubAPI.Client, account *model.Account, team *model.Team, user *model.User) error {
	// 1. 取得绑定的仓库列表
	var repositories []model.Repository
	if err := dao.DB.Table("repositories").Joins("INNER JOIN team_repositories ON (repositories.id = team_repositories.repositoriy_id AND team_id =?)", team.ID).
		Find(repositories).Error; err != nil {
		return err
	}
	// 2. 挨个仓库添加 Collaborator
	for i := 0; i < len(repositories); i++ {
		if err := AddEmployeeToRepository(ctx, client, account, &repositories[i], user); err != nil {
			return err
		}
	}
	return nil
}

// RemoveEmployeeFromTeam ..
func RemoveEmployeeFromTeam(ctx context.Context, client *GitHubAPI.Client, account *model.Account, team *model.Team, user *model.User) error {
	// 1. 取得绑定的仓库列表
	var repositories []model.Repository
	if err := dao.DB.Table("repositories").Joins("INNER JOIN team_repositories ON (repositories.id = team_repositories.repositoriy_id AND team_id =?)", team.ID).
		Find(repositories).Error; err != nil {
		return err
	}
	// 2. 挨个仓库删除 Collaborator
	for i := 0; i < len(repositories); i++ {
		if err := RemoveEmployeeFromRepository(ctx, client, account, &repositories[i], user); err != nil {
			return err
		}
	}
	return nil
}

// AddEmployeeToRepository ..
func AddEmployeeToRepository(ctx context.Context, client *GitHubAPI.Client, account *model.Account, repository *model.Repository, user *model.User) error {
	if _, err := client.Repositories.AddCollaborator(ctx, account.Login, repository.Name, user.Login, nil); err != nil {
		return err
	}
	return RepositorySync(ctx, client, account, repository)
}

// RemoveEmployeeFromRepository ..
func RemoveEmployeeFromRepository(ctx context.Context, client *GitHubAPI.Client, account *model.Account, repository *model.Repository, user *model.User) error {
	var ur model.UserRepository
	if err := dao.DB.Where("user_id = ? AND repository_id = ?", user.ID, repository.ID).First(&ur).Error; err != nil {
		return err
	}
	if ur.InvitationID > 0 {
		if _, err := client.Repositories.DeleteInvitation(ctx, account.Login, repository.Name, ur.InvitationID); err != nil {
			return err
		}
	} else {
		if _, err := client.Repositories.RemoveCollaborator(ctx, account.Login, repository.Name, user.Login); err != nil {
			return err
		}
	}
	return nil
}
