package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/naiba/poorsquad/model"
)

var cfg *model.Config
var db *gorm.DB

// RunWeb ..
func RunWeb(cf *model.Config, d *gorm.DB) {
	cfg = cf
	db = d
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"tf": func(t time.Time) string {
			return t.Format("2006年1月2号")
		},
	})
	r.Static("/static", "resource/static")
	r.LoadHTMLGlob("resource/template/**/*")

	guestPage := r.Group("")
	{
		guestPage.Use(authorize(authorizeOption{
			Guest:    true,
			IsPage:   true,
			Msg:      "您已登录",
			Btn:      "返回首页",
			Redirect: "/",
		}))
		ServeOauth2(guestPage, cf)
		guestPage.GET("/login", login)
	}

	memberPage := r.Group("")
	{
		memberPage.Use(authorize(authorizeOption{
			Member:   true,
			IsPage:   true,
			Msg:      "此页面需要登录",
			Btn:      "点此登录",
			Redirect: "/login",
		}))
		memberPage.GET("/", home)
		memberPage.GET("/company/:id/account", account)
	}

	api := r.Group("api")
	{
		memberAPI := api.Group("")
		{
			memberAPI.Use(authorize(authorizeOption{
				Member:   true,
				IsPage:   false,
				Msg:      "此页面需要登录",
				Btn:      "点此登录",
				Redirect: "/login",
			}))
			ServeCompany(memberAPI)
			memberAPI.POST("/logout", logout)
		}
	}

	r.Run()
}

type errInfo struct {
	Code  uint64
	Title string
	Msg   string
	Link  string
	Btn   string
}

func showErrorPage(c *gin.Context, i errInfo, isPage bool) {
	if isPage {
		c.HTML(http.StatusOK, "page/error", commonEnvironment(c, gin.H{
			"Code":  i.Code,
			"Title": i.Title,
			"Msg":   i.Msg,
			"Link":  i.Link,
			"Btn":   i.Btn,
		}))
	} else {
		c.JSON(http.StatusOK, model.Response{
			Code:    i.Code,
			Message: i.Msg,
		})
	}
	c.Abort()
}

func commonEnvironment(c *gin.Context, data map[string]interface{}) gin.H {
	// 站点标题
	if t, has := data["Title"]; !has {
		data["Title"] = cfg.Site.Brand
	} else {
		data["Title"] = fmt.Sprintf("%s - %s", t, cfg.Site.Brand)
	}
	u, ok := c.Get(model.CtxKeyAuthorizedUser)
	if ok {
		data["User"] = u
	}
	return data
}

type authorizeOption struct {
	Guest    bool
	Member   bool
	IsPage   bool
	Msg      string
	Redirect string
	Btn      string
}

func authorize(opt authorizeOption) func(*gin.Context) {
	return func(c *gin.Context) {
		token, err := c.Cookie(cfg.Site.CookieName)
		var code uint64 = http.StatusForbidden
		if opt.Guest {
			code = http.StatusBadRequest
		}
		commonErr := errInfo{
			Title: "访问受限",
			Code:  code,
			Msg:   opt.Msg,
			Link:  opt.Redirect,
			Btn:   opt.Btn,
		}
		if err == nil {
			var u model.User
			err = db.Where("token = ?", token).First(&u).Error
			if err == nil {
				// 已登录且只能游客访问
				if opt.Guest {
					showErrorPage(c, commonErr, opt.IsPage)
					return
				}
				c.Set(model.CtxKeyAuthorizedUser, &u)
				return
			}
		}
		// 未登录且需要登录
		if err != nil && opt.Member {
			showErrorPage(c, commonErr, opt.IsPage)
		}
	}
}
