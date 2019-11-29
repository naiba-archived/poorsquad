package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
	GitHubService "github.com/naiba/poorsquad/service/github"
)

// AccountController ..
type AccountController struct {
}

// ServeAccount ..
func ServeAccount(r gin.IRoutes) {
	ac := AccountController{}
	r.POST("/account", ac.addOrEdit)
}

type accountForm struct {
	CompanyID uint64 `binding:"required" json:"company_id,omitempty"`
	Token     string `binding:"required" json:"token,omitempty"`
}

func (ac *AccountController) addOrEdit(c *gin.Context) {
	var af accountForm
	if err := c.ShouldBindJSON(&af); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("格式错误：%s", err),
		})
		return
	}
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)

	// 验证管理权限
	var comp model.Company
	comp.ID = af.CompanyID
	if _, err := comp.CheckUserPermission(dao.DB, u.ID, model.UCPMember); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	ctx := context.Background()
	client := GitHubService.NewAPIClient(ctx, af.Token)
	gu, _, err := client.Users.Get(ctx, "")
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("验证Token失败：%s", err),
		})
		return
	}
	a := model.NewAccountFromGitHub(gu)
	a.Token = af.Token
	a.CompanyID = af.CompanyID
	if err := dao.DB.Save(&a).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	go GitHubService.AccountSync(ctx, client, &a)

	c.JSON(http.StatusOK, model.Response{
		Code:   http.StatusOK,
		Result: a,
	})
}
