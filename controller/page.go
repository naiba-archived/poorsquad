package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/naiba/poorsquad/model"
)

func login(c *gin.Context) {
	c.HTML(http.StatusOK, "user/login", commonEnvironment(c, gin.H{
		"Title": "登录",
	}))
}

func home(c *gin.Context) {
	var companies []model.Company
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	db.Table("companies").Joins("INNER JOIN user_companies ON (companies.id = user_companies.company_id AND user_companies.user_id = ?)", u.ID).Find(&companies)
	c.HTML(http.StatusOK, "page/home", commonEnvironment(c, gin.H{
		"Companies": companies,
	}))
}

func company(c *gin.Context) {
	compID := c.Param("id")
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	var comp model.Company
	if err := db.Table("companies").Joins("INNER JOIN user_companies ON (companies.id = user_companies.company_id AND user_companies.user_id = ? AND user_companies.company_id = ?)", u.ID, compID).First(&comp).Error; err != nil {
		showErrorPage(c, errInfo{
			Code:  http.StatusForbidden,
			Title: "访问受限",
			Msg:   fmt.Sprintf("无权访问：%s", err),
			Link:  "/",
			Btn:   "返回首页",
		}, true)
		return
	}

	var accounts []model.Account
	db.Where("company_id = ? ", compID).Find(&accounts)
	var teams []model.Team
	db.Where("company_id = ? ", compID).Find(&teams)

	c.HTML(http.StatusOK, "page/company", commonEnvironment(c, gin.H{
		"Title":     comp.Brand + "- 企业",
		"Company":   comp,
		"Teams":     teams,
		"Accounts":  accounts,
		"CompanyID": compID,
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
			Message: "用户ID不匹配",
		})
		return
	}
	u.Token = ""
	u.TokenExpired = time.Now()
	if err := db.Save(u).Error; err != nil {
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
