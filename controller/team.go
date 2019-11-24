package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/naiba/poorsquad/model"
)

// TeamController ..
type TeamController struct {
}

// ServeTeam ..
func ServeTeam(r gin.IRoutes) {
	tc := TeamController{}
	r.POST("/team", tc.addOrEdit)
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

	var uc model.UserCompany
	if err := db.Where("user_id = ? AND company_id = ?", u.ID, tf.CompanyID).First(&uc).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("您不是该企业的雇员：%s", err),
		})
		return
	}

	if uc.Permission < model.UCPManager {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: "您不是该企业的管理人员",
		})
		return
	}

	var t model.Team
	t.Name = tf.Name
	t.CompanyID = tf.CompanyID
	if err := db.Save(&t).Error; err != nil {
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
