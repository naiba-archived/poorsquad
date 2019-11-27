package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
)

// EmployeeController ..
type EmployeeController struct {
}

// ServeEmployee ..
func ServeEmployee(r gin.IRoutes) {
	ec := EmployeeController{}
	r.POST("/employee", ec.addOrEdit)
}

type employeeForm struct {
	Type       string `binding:"required" json:"type,omitempty"`
	ID         uint64 `binding:"required,min=1" json:"id,omitempty"`
	Username   string `binding:"required" json:"username,omitempty"`
	Permission uint64 `json:"permission,omitempty"`
}

func (ec *EmployeeController) addOrEdit(c *gin.Context) {
	var ef employeeForm
	if err := c.ShouldBindJSON(&ef); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("格式错误：%s", err),
		})
		return
	}
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)

	var company model.Company
	var team model.Team
	var user model.User
	var repository model.Repository
	var teamsID []uint64
	var err error

	switch ef.Type {
	case "company":
		err = dao.DB.Where("id = ?", ef.ID).First(&company).Error
	case "team":
		err = dao.DB.Where("id = ?", ef.ID).First(&team).Error
		if err == nil {
			err = dao.DB.Where("id = ?", team.CompanyID).First(&company).Error
		}
	case "repository":
		err = dao.DB.Where("id = ?", ef.ID).First(&repository).Error
		if err == nil {
			var teamRepositories []model.TeamRepository
			err = dao.DB.Where("repository_id = ?", repository.ID).Find(&teamRepositories).Error
			if err == nil && len(teamRepositories) > 0 {
				for i := 0; i < len(teamRepositories); i++ {
					teamsID = append(teamsID, teamRepositories[i].TeamID)
				}
			}
			var account model.Account
			err = dao.DB.Where("id = ?", repository.AccountID).First(&account).Error
			if err == nil {
				err = dao.DB.Where("id = ?", account.CompanyID).First(&company).Error
			}
		}
	default:
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("错误：%s", "不支持的操作"),
		})
		return
	}

	if err == nil {
		err = dao.DB.Where("login = ?", ef.Username).First(&user).Error
	}

	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	var respData interface{}

	// 验证管理权限
	companyPerm, errCompanyAdmin := company.CheckUserPermission(dao.DB, u.ID, model.UCPMember)
	teamPerm, errTeamAdmin := team.CheckUserPermission(dao.DB, u.ID, model.UTPMember)
	switch ef.Type {
	case "company":
		if errCompanyAdmin != nil {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("访问受限：%s", errCompanyAdmin),
			})
			return
		}
		if companyPerm < ef.Permission {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("访问受限：%s", "授权不能高于您自身权限"),
			})
			return
		}
		var userCompany model.UserCompany
		userCompany.CompanyID = company.ID
		userCompany.UserID = user.ID
		userCompany.Permission = ef.Permission
		err = dao.DB.Save(&userCompany).Error
		respData = userCompany
	case "team":
		if errCompanyAdmin != nil && errTeamAdmin != nil {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("访问受限：%s", errTeamAdmin),
			})
			return
		}
		if companyPerm < model.UCPManager && (teamPerm < ef.Permission || teamPerm < model.UTPManager) {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("访问受限：%s", "授权不能高于您自身权限"),
			})
			return
		}
		var userTeam model.UserTeam
		userTeam.TeamID = team.ID
		userTeam.UserID = user.ID
		userTeam.Permission = ef.Permission
		err = dao.DB.Save(&userTeam).Error
		respData = userTeam
	case "repository":
		if errTeamAdmin != nil && errCompanyAdmin != nil {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("访问受限：%s", errTeamAdmin),
			})
			return
		}

		if companyPerm < model.UCPManager && teamPerm < model.UTPManager {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("访问受限：%s", "您没有权限添加外部雇员"),
			})
			return
		}

		var userReposity model.UserRepository
		userReposity.RepositoryID = repository.ID
		userReposity.UserID = user.ID
		err = dao.DB.Save(&userReposity).Error
		respData = userReposity
	}

	c.JSON(http.StatusOK, model.Response{
		Code:   http.StatusOK,
		Result: respData,
	})
}

//TODO: 雇员增减跟 GitHub 互通
