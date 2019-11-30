package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
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

	// 仓库列表
	var repos []model.Repository
	dao.DB.Where("account_id IN (?) ", accountID).Find(&repos)
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

	// 管理员列表
	dao.FetchCompanyManagers(&comp)

	c.HTML(http.StatusOK, "page/company", commonEnvironment(c, gin.H{
		"Title":        comp.Brand + "- 企业",
		"Company":      comp,
		"Teams":        teams,
		"Repositories": repos,
		"Accounts":     accounts,
		"CompanyID":    compID,
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
	u.Token = ""
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
