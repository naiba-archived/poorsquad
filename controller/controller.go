package controller

import (
	"fmt"
	"net/http"

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
	r.Static("/static", "resource/static")
	r.LoadHTMLGlob("resource/template/**/*")
	guestPage := r.Use(authorize(authorizeOption{
		Guest:  true,
		IsPage: true,
	}))
	guestPage.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "user/login", gin.H{})
	})
	ServeOauth2(guestPage, cf)
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
		if i.Btn == "" {
			i.Btn = "返回首页"
		}
		c.HTML(http.StatusOK, "page/error", commonEnvironment(gin.H{
			"Code":  i.Code,
			"Title": i.Title,
			"Msg":   i.Msg,
			"Link":  i.Link,
		}))
	} else {
		c.JSON(http.StatusOK, model.Response{
			Code:    i.Code,
			Message: i.Msg,
		})
	}
	c.Abort()
}

func commonEnvironment(data map[string]interface{}) gin.H {
	// 站点标题
	if data["Title"] == "" {
		data["Title"] = cfg.Site.Brand
	} else {
		data["Title"] = fmt.Sprintf("%s - %s", data["Title"], cfg.Site.Brand)
	}
	return data
}

type authorizeOption struct {
	Guest    bool
	Member   bool
	IsPage   bool
	Redirect string
	Btn      string
}

func authorize(opt authorizeOption) func(*gin.Context) {
	return func(c *gin.Context) {
		token, err := c.Cookie(cfg.Site.CookieName)
		if err != nil && opt.Member {
			showErrorPage(c, errInfo{}, opt.IsPage)
			return
		}
		// TODO: authorize user
		if err == nil {
			db.Where("token = ?", token)
		}
	}
}
