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
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "user/login", gin.H{})
	})
	ServeOauth2(r, cf)
	r.Run()
}

type errInfo struct {
	Code  uint64
	Title string
	Msg   string
	Link  string
}

func showErrorPage(c *gin.Context, i errInfo) {
	c.HTML(http.StatusOK, "page/error", commonEnvironment(gin.H{
		"Code":  i.Code,
		"Title": i.Title,
		"Msg":   i.Msg,
		"Link":  i.Link,
	}))
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
