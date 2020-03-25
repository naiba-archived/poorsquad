package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	GitHubAPI "github.com/google/go-github/v28/github"
	"github.com/jinzhu/gorm"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
	"github.com/naiba/poorsquad/service/github"
	GitHubService "github.com/naiba/poorsquad/service/github"
)

// RepositoryController ..
type RepositoryController struct {
}

// ServeRepository ..
func ServeRepository(r gin.IRoutes) {
	rc := RepositoryController{}
	r.POST("/repository", rc.addOrEdit)
	r.POST("/webhook", rc.addOrEditWebhook)
	r.DELETE("/repository/:rid/delete/:name", rc.delete)
	r.DELETE("/repository/:rid/webhook/:wid", rc.deleteWebhook)
	r.GET("/repository/:rid/webhook/:wid/ping", rc.pingWebhook)
	r.GET("/repository/:rid/webhook/:wid/test", rc.testWebhook)
}

type repositoryForm struct {
	ID        uint64 `json:"id,omitempty"`
	Name      string `binding:"required" json:"name,omitempty"`
	AccountID uint64 `binding:"required" json:"account_id,omitempty"`
	Private   string `binding:"required" json:"private,omitempty"`
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
	var distAccount model.Account
	if err := dao.DB.First(&distAccount, "id = ?", rf.AccountID).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求受限：%s", err),
		})
		return
	}

	var comp model.Company
	comp.ID = distAccount.CompanyID
	if _, err := comp.CheckUserPermission(dao.DB, u.ID, model.UCPSuperManager); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	var err error
	var repostory model.Repository
	var repo GitHubAPI.Repository
	if rf.ID != 0 {
		if rf.AccountID != distAccount.ID {
			err = errors.New("GitHub 尚未完善账户间转移 API")
		}
		if err == nil {
			repostory.ID = rf.ID
			err = dao.DB.First(&repostory).Error
		}
	}
	// 添加仓库
	ctx := context.Background()
	client := github.NewAPIClient(ctx, distAccount.Token)
	repo.Name = &rf.Name
	private := rf.Private == "on"
	repo.Private = &private
	var resp *GitHubAPI.Repository
	if rf.ID != 0 {
		resp, _, err = client.Repositories.Edit(ctx, distAccount.Login, repostory.Name, &repo)
	} else {
		resp, _, err = client.Repositories.Create(ctx, "", &repo)
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("GitHub 同步：%s", err),
		})
		return
	}
	r := model.NewRepositoryFromGitHub(resp)
	r.AccountID = distAccount.ID
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

func (rc *RepositoryController) delete(c *gin.Context) {
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)

	// 验证管理权限
	var repo model.Repository
	var account model.Account
	var comp model.Company
	err := dao.DB.First(&repo, "id = ?", c.Param("rid")).Error

	if err == nil {
		if repo.Name != c.Param("name") {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusBadRequest,
				Message: "仓库名称不匹配",
			})
			return
		}
		err = dao.DB.First(&account, "id = ?", repo.AccountID).Error
	}

	if err == nil {
		comp.ID = account.CompanyID
		_, err = comp.CheckUserPermission(dao.DB, u.ID, model.UCPSuperManager)
	}

	var tx *gorm.DB
	if err == nil {
		tx = dao.DB.Begin()
		err = tx.Delete(model.UserRepository{}, "repository_id = ?", repo.ID).Error
	}
	if err == nil {
		err = tx.Delete(repo).Error
	}
	if err == nil {
		ctx := context.Background()
		client := github.NewAPIClient(ctx, account.Token)
		_, err = client.Repositories.Delete(ctx, account.Login, repo.Name)
	}
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("出现错误：%s", err),
		})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}

type webhookForm struct {
	ID          int64  `json:"id"`
	RepoID      uint64 `json:"repo_id"`
	URL         string `binding:"required"`
	Events      string `binding:"required"`
	Secret      string `binding:"required"`
	ContentType string `binding:"required" json:"content_type"`
	Active      string
	InsecureSSL string `json:"insecure_ssl"`
}

func (rc *RepositoryController) addOrEditWebhook(c *gin.Context) {
	var err error
	var wf webhookForm
	var events []string
	if err := c.ShouldBindJSON(&wf); err != nil {

	}
	if err == nil {
		err = json.Unmarshal([]byte(wf.Events), &events)
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("数据格式不对：%s", err),
		})
		return
	}

	var repo model.Repository
	var comp model.Company
	err = dao.DB.First(&repo, "id = ?", wf.RepoID).Error
	if err == nil {
		err = repo.ReleatedAccount(dao.DB)
	}
	if err == nil {
		err = dao.DB.First(&comp, "id = ?", repo.Account.CompanyID).Error
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	var has bool
	has, _ = repo.HasUser(dao.DB, u.ID)
	if !has {
		teams, err := repo.GetTeams(dao.DB)
		if err == nil {
			has, err = u.InTeams(dao.DB, teams, model.UTPManager)
		}
	}
	if !has {
		_, err = comp.CheckUserPermission(dao.DB, u.ID, model.UCPManager)
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusForbidden,
			Message: "无权访问此仓库",
		})
		return
	}

	var hook GitHubAPI.Hook
	hook.Events = events
	var active = wf.Active == "on"
	hook.Active = &active
	hook.Config = make(map[string]interface{})
	hook.Config["url"] = wf.URL
	hook.Config["content_type"] = wf.ContentType
	hook.Config["secret"] = wf.Secret
	hook.Config["insecure_ssl"] = 0
	if wf.InsecureSSL == "on" {
		hook.Config["insecure_ssl"] = 1
	}

	ctx := context.Background()
	client := GitHubService.NewAPIClient(ctx, repo.Account.Token)
	if wf.ID > 0 {
		_, _, err = client.Repositories.EditHook(ctx, repo.Account.Login, repo.Name, wf.ID, &hook)
	} else {
		_, _, err = client.Repositories.CreateHook(ctx, repo.Account.Login, repo.Name, &hook)
	}

	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("出现错误：%s", err),
		})
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}

func (rc *RepositoryController) deleteWebhook(c *gin.Context) {
	rid, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if rid < 1 {
		err = errors.New("错误ID")
	}
	wid, _ := strconv.ParseInt(c.Param("wid"), 10, 64)
	if wid < 1 {
		err = errors.New("错误ID")
	}
	var repo model.Repository
	var comp model.Company
	if err == nil {
		err = dao.DB.First(&repo, "id = ?", rid).Error
	}
	if err == nil {
		err = repo.ReleatedAccount(dao.DB)
	}
	if err == nil {
		err = dao.DB.First(&comp, "id = ?", repo.Account.CompanyID).Error
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	var has bool
	has, _ = repo.HasUser(dao.DB, u.ID)
	if !has {
		teams, err := repo.GetTeams(dao.DB)
		if err == nil {
			has, err = u.InTeams(dao.DB, teams, model.UTPManager)
		}
	}
	if !has {
		_, err = comp.CheckUserPermission(dao.DB, u.ID, model.UCPManager)
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusForbidden,
			Message: "无权访问此仓库",
		})
		return
	}

	ctx := context.Background()
	client := GitHubService.NewAPIClient(ctx, repo.Account.Token)
	_, err = client.Repositories.DeleteHook(ctx, repo.Account.Login, repo.Name, wid)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("出现错误：%s", err),
		})
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}

func (rc *RepositoryController) pingWebhook(c *gin.Context) {
	rid, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if rid < 1 {
		err = errors.New("错误ID")
	}
	wid, _ := strconv.ParseInt(c.Param("wid"), 10, 64)
	if wid < 1 {
		err = errors.New("错误ID")
	}
	var repo model.Repository
	var comp model.Company
	err = dao.DB.First(&repo, "id = ?", rid).Error
	if err == nil {
		err = repo.ReleatedAccount(dao.DB)
	}
	if err == nil {
		err = dao.DB.First(&comp, "id = ?", repo.Account.CompanyID).Error
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	var has bool
	has, _ = repo.HasUser(dao.DB, u.ID)
	if !has {
		teams, err := repo.GetTeams(dao.DB)
		if err == nil {
			has, err = u.InTeams(dao.DB, teams, 0)
		}
	}
	if !has {
		_, err = comp.CheckUserPermission(dao.DB, u.ID, model.UCPManager)
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusForbidden,
			Message: "无权访问此仓库",
		})
		return
	}

	ctx := context.Background()
	client := GitHubService.NewAPIClient(ctx, repo.Account.Token)
	_, err = client.Repositories.PingHook(ctx, repo.Account.Login, repo.Name, wid)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("出现错误：%s", err),
		})
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}

func (rc *RepositoryController) testWebhook(c *gin.Context) {
	rid, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if rid < 1 {
		err = errors.New("错误ID")
	}
	wid, _ := strconv.ParseInt(c.Param("wid"), 10, 64)
	if wid < 1 {
		err = errors.New("错误ID")
	}
	var repo model.Repository
	var comp model.Company
	err = dao.DB.First(&repo, "id = ?", rid).Error
	if err == nil {
		err = repo.ReleatedAccount(dao.DB)
	}
	if err == nil {
		err = dao.DB.First(&comp, "id = ?", repo.Account.CompanyID).Error
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	var has bool
	has, _ = repo.HasUser(dao.DB, u.ID)
	if !has {
		teams, err := repo.GetTeams(dao.DB)
		if err == nil {
			has, err = u.InTeams(dao.DB, teams, 0)
		}
	}
	if !has {
		_, err = comp.CheckUserPermission(dao.DB, u.ID, model.UCPManager)
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusForbidden,
			Message: "无权访问此仓库",
		})
		return
	}

	ctx := context.Background()
	client := GitHubService.NewAPIClient(ctx, repo.Account.Token)
	_, err = client.Repositories.TestHook(ctx, repo.Account.Login, repo.Name, wid)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("出现错误：%s", err),
		})
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}
