package controller

import (
	"context"
	"log"
	"net/http"

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
func ServeOauth2(r *gin.Engine, cf *model.Config) {
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
		log.Fatal(err)
	}
	oc := oa.oauth2Config.Client(ctx, otk)
	client := githubApi.NewClient(oc)
	u, _, _ := client.Users.Get(ctx, "")
	c.JSON(http.StatusOK, u)
}
