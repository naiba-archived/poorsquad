package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
	"github.com/naiba/poorsquad/service/github"
)

// TeamController ..
type TeamController struct {
}

// ServeTeam ..
func ServeTeam(r gin.IRoutes) {
	tc := TeamController{}
	r.POST("/team", tc.addOrEdit)
	r.POST("/team/repositories", tc.bindRepositories)
}

type teamForm struct {
	CompanyID uint64 `binding:"required" json:"company_id,omitempty"`
	Name      string `binding:"required" json:"name,omitempty"`
}

func (tc *TeamController) addOrEdit(c *gin.Context) {
	var tf teamForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("格式错误：%s", err),
		})
		return
	}
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)

	// 验证管理权限
	var comp model.Company
	comp.ID = tf.CompanyID
	if _, err := comp.CheckUserPermission(dao.DB, u.ID, model.UCPMember); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	var t model.Team
	t.Name = tf.Name
	t.CompanyID = tf.CompanyID
	if err := dao.DB.Save(&t).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:   http.StatusOK,
		Result: t,
	})
}

type teamRepositoriesRequest struct {
	ID             uint64   `binding:"required" json:"id,omitempty"`
	RepositoriesID []uint64 `json:"repositories_id,omitempty"`
}

func (tc *TeamController) bindRepositories(c *gin.Context) {
	var trr teamRepositoriesRequest
	if err := c.ShouldBindJSON(&trr); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求参数错误：%s", err),
		})
		return
	}

	// 验证小组是否存在
	var t model.Team
	if err := dao.DB.Where("id = ?", trr.ID).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求受限：%s", err),
		})
		return
	}

	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	// 权限验证
	var company model.Company
	company.ID = t.CompanyID
	_, err := company.CheckUserPermission(dao.DB, u.ID, model.UCPMember)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求受限：%s", err),
		})
		return
	}
	_, err = t.CheckUserPermission(dao.DB, u.ID, model.UTPMember)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求受限：%s", err),
		})
		return
	}

	// 验证仓库是否存在，并属于此企业
	var count int
	dao.DB.Table("repositories").Where("account_id IN (SELECT accounts.id FROM accounts WHERE company_id = ?) AND id IN (?)", t.CompanyID, trr.RepositoriesID).Count(&count)
	if count != len(trr.RepositoriesID) {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求受限：%s", "请检查仓库列表中的仓库是否属于本公司"),
		})
		return
	}

	var trs []model.TeamRepository
	if err := dao.DB.Where("team_id = ?", t.ID).Find(&trs).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	// 1. 要清理的仓库
CHECKDEL:
	for i := 0; i < len(trs); i++ {
		tr := trs[i]
		for j := 0; j < len(trr.RepositoriesID); j++ {
			repoID := trr.RepositoriesID[j]
			if tr.RepositoryID == repoID {
				continue CHECKDEL
			}
		}
		github.RemoveRepositoryFromTeam(&t, tr.RepositoryID)
		dao.DB.Delete(&tr)
	}
	// 2. 要添加的仓库
CHECKADD:
	for j := 0; j < len(trr.RepositoriesID); j++ {
		repoID := trr.RepositoriesID[j]
		for i := 0; i < len(trs); i++ {
			tr := trs[i]
			if tr.RepositoryID == repoID {
				continue CHECKADD
			}
		}
		github.AddRepositoryFromTeam(&t, repoID)
		dao.DB.Save(&model.TeamRepository{
			TeamID:       t.ID,
			RepositoryID: repoID,
		})
	}

	c.JSON(http.StatusOK, model.Response{
		Code:   http.StatusOK,
		Result: trr.RepositoriesID,
	})
}
