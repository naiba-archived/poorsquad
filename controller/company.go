package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
)

// CompanyController ..
type CompanyController struct {
}

// ServeCompany ..
func ServeCompany(r gin.IRoutes) {
	cc := CompanyController{}
	r.POST("/company", cc.addOrEdit)
}

type companyForm struct {
	ID        uint64 `json:"id,omitempty"`
	Brand     string `binding:"required" json:"brand,omitempty"`
	AvatarURL string `binding:"required" json:"avatar_url,omitempty"`
}

func (cc *CompanyController) addOrEdit(c *gin.Context) {
	var cf companyForm
	if err := c.ShouldBindJSON(&cf); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("格式错误：%s", err),
		})
		return
	}
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)

	var company model.Company
	var initCompany bool
	if cf.ID != 0 {
		// 验证管理权限
		company.ID = cf.ID
		if _, err := company.CheckUserPermission(dao.DB, u.ID, model.UCPManager); err != nil {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
	} else {
		initCompany = true
	}
	company.Brand = cf.Brand
	company.AvatarURL = cf.AvatarURL

	if err := dao.DB.Save(&company).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	if initCompany {
		var uc model.UserCompany
		uc.UserID = u.ID
		uc.CompanyID = company.ID
		uc.Permission = model.UCPSuperManager
		if err := dao.DB.Save(&uc).Error; err != nil {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("数据库错误：%s", err),
			})
			return
		}
	}

	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}
