package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/naiba/poorsquad/model"
)

// RunWeb ..
func RunWeb(cf *model.Config) {
	r := gin.Default()
	r.Static("/static", "resource/static")
	r.LoadHTMLGlob("resource/template/**/*")
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "user/login", gin.H{})
	})
	ServeOauth2(r, cf)
	r.Run()
}
