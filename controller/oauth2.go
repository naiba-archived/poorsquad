package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/gin-gonic/gin"
	GitHubAPI "github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
)

// Oauth2 ..
type Oauth2 struct {
	oauth2Config *oauth2.Config
}

// ServeOauth2 ..
func ServeOauth2(r gin.IRoutes) {
	oa := Oauth2{
		oauth2Config: &oauth2.Config{
			ClientID:     dao.Conf.GitHub.ClientID,
			ClientSecret: dao.Conf.GitHub.ClientSecret,
			Scopes:       []string{},
			Endpoint:     github.Endpoint,
		},
	}
	r.GET("/oauth2/login", oa.login)
	r.GET("/oauth2/callback", oa.callback)
}

func (oa *Oauth2) login(c *gin.Context) {
	url := oa.oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

func (oa *Oauth2) callback(c *gin.Context) {
	ctx := context.Background()
	otk, err := oa.oauth2Config.Exchange(ctx, c.Query("code"))
	if err != nil {
		showErrorPage(c, errInfo{
			Code:  http.StatusBadRequest,
			Title: "登录失败",
			Msg:   fmt.Sprintf("错误信息：%s", err),
		}, true)
		return
	}
	oc := oa.oauth2Config.Client(ctx, otk)
	client := GitHubAPI.NewClient(oc)
	gu, _, err := client.Users.Get(ctx, "")
	if err != nil {
		showErrorPage(c, errInfo{
			Code:  http.StatusBadRequest,
			Title: "登录失败",
			Msg:   fmt.Sprintf("错误信息：%s", err),
		}, true)
		return
	}
	var u model.User
	err = dao.DB.Where("id = ?", gu.GetID()).First(&u).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		showErrorPage(c, errInfo{
			Code:  http.StatusBadRequest,
			Title: "登录失败",
			Msg:   fmt.Sprintf("错误信息：%s", err),
		}, true)
		return
	}
	if err == gorm.ErrRecordNotFound {
		var count uint64
		err = dao.DB.Model(&model.User{}).Count(&count).Error
		if err != nil {
			showErrorPage(c, errInfo{
				Code:  http.StatusBadRequest,
				Title: "系统错误",
				Msg:   fmt.Sprintf("错误信息：%s", err),
			}, true)
			return
		}
		u = model.NewUserFromGitHub(gu)
		if u.ID == 0 {
			showErrorPage(c, errInfo{
				Code:  http.StatusBadRequest,
				Title: "系统错误",
				Msg:   fmt.Sprintf("错误信息：%s", "未能获取到雇员ID"),
			}, true)
			return
		}
		if u.Login == "" {
			showErrorPage(c, errInfo{
				Code:  http.StatusBadRequest,
				Title: "系统错误",
				Msg:   fmt.Sprintf("错误信息：%s", "未能获取到雇员登录名"),
			}, true)
			return
		}
		u.SuperAdmin = count == 0
	}
	u.IssueNewToken()
	err = dao.DB.Save(&u).Error
	if err != nil {
		showErrorPage(c, errInfo{
			Code:  http.StatusBadRequest,
			Title: "系统错误",
			Msg:   fmt.Sprintf("雇员保存失败：%s", err),
		}, true)
		return
	}
	c.SetCookie(dao.Conf.Site.CookieName, u.Token, 60*60*24*14, "", "", false, false)
	c.Status(http.StatusOK)
	c.Writer.WriteString("<script>window.location.href='/'</script>")
}
