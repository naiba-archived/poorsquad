package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/gin-gonic/gin"
	githubApi "github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/naiba/poorsquad/model"
)

// Oauth2 ..
type Oauth2 struct {
	oauth2Config *oauth2.Config
	systemConfig *model.Config
}

// ServeOauth2 ..
func ServeOauth2(r gin.IRoutes, cf *model.Config) {
	oa := Oauth2{
		oauth2Config: &oauth2.Config{
			ClientID:     cf.GitHub.ClientID,
			ClientSecret: cf.GitHub.ClientSecret,
			Scopes:       []string{},
			Endpoint:     github.Endpoint,
		},
		systemConfig: cf,
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
	client := githubApi.NewClient(oc)
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
	err = db.Where("id = ?", gu.GetID()).First(&u).Error
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
		err = db.Model(&model.User{}).Count(&count).Error
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
				Msg:   fmt.Sprintf("错误信息：%s", "未能获取到用户ID"),
			}, true)
			return
		}
		if u.Login == "" {
			showErrorPage(c, errInfo{
				Code:  http.StatusBadRequest,
				Title: "系统错误",
				Msg:   fmt.Sprintf("错误信息：%s", "未能获取到用户登录名"),
			}, true)
			return
		}
		u.SuperAdmin = count == 0
	}
	u.IssueNewToken()
	err = db.Save(&u).Error
	if err != nil {
		showErrorPage(c, errInfo{
			Code:  http.StatusBadRequest,
			Title: "系统错误",
			Msg:   fmt.Sprintf("用户保存失败：%s", err),
		}, true)
		return
	}
	c.SetCookie(cfg.Site.CookieName, u.Token, 60*60*24*14, "", "", false, false)
	c.Status(http.StatusOK)
	c.Writer.WriteString("<script>window.location.href='/'</script>")
}