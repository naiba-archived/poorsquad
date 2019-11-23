package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/naiba/poorsquad/model"
)

// CompanyController ..
type CompanyController struct {
}

// ServeCompany ..
func ServeCompany(r gin.IRoutes) {
	cc := CompanyController{}
	r.POST("/company", cc.addOrEditCompany)
}

type companyForm struct {
	ID        uint64 `json:"id,omitempty"`
	Brand     string `binding:"required" json:"brand,omitempty"`
	AvatarURL string `binding:"required" json:"avatar_url,omitempty"`
}

func (cc *CompanyController) addOrEditCompany(c *gin.Context) {
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
	if cf.ID != 0 {
		if err := db.Where("id = ? AND user_id = ?", cf.ID, u.ID).First(&company).Error; err != nil {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("未找到此公司：%s", err),
			})
			return
		}
	} else {
		company.UserID = u.ID
	}
	company.Brand = cf.Brand
	company.AvatarURL = cf.AvatarURL

	if err := db.Save(&company).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	time.Sleep(time.Second * 2)

	c.JSON(http.StatusOK, model.Response{
		Code:   http.StatusOK,
		Result: company,
	})
}
