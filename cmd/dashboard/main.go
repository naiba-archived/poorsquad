package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Static("/static", "resource/static")
	r.LoadHTMLGlob("resource/template/**/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "user/login", gin.H{})
	})
	r.Run()
}
