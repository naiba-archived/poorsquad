package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v28/github"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
	GitHubService "github.com/naiba/poorsquad/service/github"
)

func login(c *gin.Context) {
	c.HTML(http.StatusOK, "user/login", commonEnvironment(c, gin.H{
		"Title": "登录",
	}))
}

func home(c *gin.Context) {
	var companies []model.Company
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	dao.DB.Raw(`SELECT "companies".* FROM "companies" INNER JOIN user_companies ON (companies.id = user_companies.company_id AND user_companies.user_id = ?) WHERE "companies"."deleted_at" IS NULL
	UNION SELECT "companies".* FROM "companies" INNER JOIN user_teams,teams ON (companies.id = teams.company_id AND teams.id = user_teams.team_id AND user_teams.user_id = ?) WHERE "companies"."deleted_at" IS NULL`, u.ID, u.ID).Scan(&companies)
	for i := 0; i < len(companies); i++ {
		// 管理员列表
		dao.FetchCompanyManagers(&companies[i])
	}
	c.HTML(http.StatusOK, "page/home", commonEnvironment(c, gin.H{
		"Companies": companies,
	}))
}

func company(c *gin.Context) {
	compID := c.Param("id")
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	var comp model.Company
	if err := dao.DB.Raw(`SELECT "companies".* FROM "companies" INNER JOIN user_companies ON (companies.id = user_companies.company_id AND user_companies.user_id = ?) WHERE "companies"."deleted_at" IS NULL AND companies.id = ?
	UNION SELECT "companies".* FROM "companies" INNER JOIN user_teams,teams ON (companies.id = teams.company_id AND teams.id = user_teams.team_id AND user_teams.user_id = ?) WHERE "companies"."deleted_at" IS NULL AND companies.id = ?`, u.ID, compID, u.ID, compID).Scan(&comp).Error; err != nil {
		showErrorPage(c, errInfo{
			Code:  http.StatusForbidden,
			Title: "访问受限",
			Msg:   fmt.Sprintf("无权访问：%s", err),
			Link:  "/",
			Btn:   "返回首页",
		}, true)
		return
	}

	// 账号列表
	var accounts []model.Account
	dao.DB.Where("company_id = ? ", compID).Find(&accounts)
	var accountID []uint64
	for i := 0; i < len(accounts); i++ {
		accountID = append(accountID, accounts[i].ID)
	}

	// 管理员列表
	dao.FetchCompanyManagers(&comp)

	// 仓库列表
	var repos []model.Repository
	if _, err := comp.CheckUserPermission(dao.DB, u.ID, model.UCPManager); err == nil {
		dao.DB.Where("account_id IN (?) ", accountID).Find(&repos)
	} else {
		dao.DB.Raw(`SELECT  repositories.* FROM repositories INNER JOIN accounts,user_repositories WHERE repositories.id = user_repositories.repository_id  AND repositories.account_id = accounts.id AND accounts.company_id = ? AND user_repositories.user_id = ?`, comp.ID, u.ID).Scan(&repos)
	}
	// 外部雇员
	for i := 0; i < len(repos); i++ {
		repos[i].RelatedOutsideCollaborators(dao.DB)
		// 拉取外部雇员详细信息
		for j := 0; j < len(repos[i].OutsideCollaborators); j++ {
			user, err := dao.GetUserByID(repos[i].OutsideCollaborators[j].ID)
			if err != nil {
				showErrorPage(c, errInfo{
					Code:  http.StatusInternalServerError,
					Title: "获取雇员信息错误",
					Msg:   fmt.Sprintf("缓存错误：%s", err),
					Link:  "/",
					Btn:   "返回首页",
				}, true)
				return
			}
			repos[i].OutsideCollaborators[j] = user
		}
	}

	var teams []model.Team
	dao.DB.Where("company_id = ? ", compID).Find(&teams)
	for i := 0; i < len(teams); i++ {
		teams[i].FetchRepositoriesID(dao.DB)
		teams[i].FetchEmployeesID(dao.DB)
		teams[i].FetchEmployees(dao.DB)
		for j := 0; j < len(teams[i].Managers); j++ {
			user, _ := dao.GetUserByID(teams[i].Managers[j].ID)
			teams[i].Managers[j] = user
		}
		for j := 0; j < len(teams[i].Employees); j++ {
			user, _ := dao.GetUserByID(teams[i].Employees[j].ID)
			teams[i].Employees[j] = user
		}
	}

	c.HTML(http.StatusOK, "page/company", commonEnvironment(c, gin.H{
		"Title":        comp.Brand + "- 企业",
		"Company":      comp,
		"Teams":        teams,
		"Repositories": repos,
		"Accounts":     accounts,
		"CompanyID":    compID,
	}))
}

func repository(c *gin.Context) {
	repoID := c.Param("id")
	var repo model.Repository
	var comp model.Company
	var err error
	err = dao.DB.First(&repo, "id = ?", repoID).Error
	if err == nil {
		err = repo.ReleatedAccount(dao.DB)
	}
	if err == nil {
		err = dao.DB.First(&comp, "id = ?", repo.Account.CompanyID).Error
	}
	if err != nil {
		showErrorPage(c, errInfo{
			Code:  http.StatusForbidden,
			Title: "出现错误",
			Msg:   fmt.Sprintf("数据库错误：%s", err),
			Link:  "/",
			Btn:   "返回首页",
		}, true)
		return
	}

	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	var has bool
	has, _ = repo.HasUser(dao.DB, u.ID)
	if !has {
		teams, err := repo.GetTeams(dao.DB)
		if err == nil {
			has, err = u.InTeams(dao.DB, teams)
		}
	}
	if !has {
		_, err = comp.CheckUserPermission(dao.DB, u.ID, model.UCPManager)
	}
	if err != nil {
		showErrorPage(c, errInfo{
			Code:  http.StatusForbidden,
			Title: "出现错误",
			Msg:   "无权访问此仓库",
			Link:  "/",
			Btn:   "返回首页",
		}, true)
		return
	}

	ctx := context.Background()
	client := GitHubService.NewAPIClient(ctx, repo.Account.Token)

	var allHooks []*github.Hook
	nextPage := 1
	for nextPage != 0 {
		hooks, resp, err := client.Repositories.ListHooks(ctx, repo.Account.Login, repo.Name, &github.ListOptions{
			Page: nextPage,
		})
		if err != nil {
			showErrorPage(c, errInfo{
				Code:  http.StatusForbidden,
				Title: "出现错误",
				Msg:   "GitHub API：" + err.Error(),
				Link:  "/",
				Btn:   "返回首页",
			}, true)
			return
		}
		nextPage = resp.NextPage
		allHooks = append(allHooks, hooks...)
	}

	c.HTML(http.StatusOK, "page/repository", commonEnvironment(c, gin.H{
		"Title":   repo.Name + "- 仓库",
		"Hooks":   allHooks,
		"Company": comp,
	}))
}

type logoutForm struct {
	ID uint64
}

func logout(c *gin.Context) {
	var lf logoutForm
	if err := c.ShouldBindJSON(&lf); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求错误：%s", err),
		})
		return
	}
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	if u.ID != lf.ID {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: "雇员ID不匹配",
		})
		return
	}
	u.TokenExpired = time.Now()
	if err := dao.DB.Save(u).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}
