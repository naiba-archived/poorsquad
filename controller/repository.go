package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	GitHubAPI "github.com/google/go-github/v28/github"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
	"github.com/naiba/poorsquad/service/github"
)

// RepositoryController ..
type RepositoryController struct {
}

// ServeRepository ..
func ServeRepository(r gin.IRoutes) {
	rc := RepositoryController{}
	r.POST("/repository", rc.addOrEdit)
}

type repositoryForm struct {
	Name    string `binding:"required"`
	Account string `binding:"required"`
	Private string `binding:"required"`
}

func (rc *RepositoryController) addOrEdit(c *gin.Context) {
	var rf repositoryForm
	if err := c.ShouldBindJSON(&rf); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("格式错误：%s", err),
		})
		return
	}
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)

	// 验证管理权限
	var account model.Account
	if err := dao.DB.First(&account, "id = ?", rf.Account).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求受限：%s", err),
		})
		return
	}
	var comp model.Company
	comp.ID = account.CompanyID
	if _, err := comp.CheckUserPermission(dao.DB, u.ID, model.UCPManager); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// 添加仓库
	ctx := context.Background()
	client := github.NewAPIClient(ctx, account.Token)
	var repo GitHubAPI.Repository
	repo.Name = &rf.Name
	private := rf.Private == "on"
	repo.Private = &private
	resp, _, err := client.Repositories.Create(ctx, "", &repo)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("GitHub 同步：%s", err),
		})
		return
	}
	r := model.NewRepositoryFromGitHub(resp)
	r.AccountID = account.ID
	r.SyncedAt = time.Now()
	if err := dao.DB.Save(&r).Error; err != nil {
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
