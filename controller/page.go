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
	dao.DB.Table("companies").Joins("INNER JOIN user_companies ON (companies.id = user_companies.company_id AND user_companies.user_id = ?)", u.ID).Find(&companies)
	c.HTML(http.StatusOK, "page/home", commonEnvironment(c, gin.H{
		"Companies": companies,
	}))
}

func company(c *gin.Context) {
	compID := c.Param("id")
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	var comp model.Company
	if err := dao.DB.Table("companies").Joins("INNER JOIN user_companies ON (companies.id = user_companies.company_id AND user_companies.user_id = ? AND user_companies.company_id = ?)", u.ID, compID).First(&comp).Error; err != nil {
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
		repos[i].RelatedOutsideUsers(dao.DB)
		// 拉取外部雇员详细信息
		for j := 0; j < len(repos[i].OutsideUsers); j++ {
			user, err := dao.GetUserByID(repos[i].OutsideUsers[j].ID)
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
			repos[i].OutsideUsers[j] = user
		}
	}

	var teams []model.Team
	dao.DB.Where("company_id = ? ", compID).Find(&teams)

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
