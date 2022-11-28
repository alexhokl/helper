package authhelper

import (
	"fmt"

	"golang.org/x/oauth2"
)

type OAuthConfig struct {
	ClientId     string
	ClientSecret string
	Endpoint     oauth2.Endpoint
	Scopes       []string
	RedirectURI  string
	Port         int
}

func (c *OAuthConfig) GetOAuthConfig() *oauth2.Config {
	config := &oauth2.Config{
		ClientID:     c.ClientId,
		ClientSecret: c.ClientSecret,
		Scopes:       c.Scopes,
		Endpoint:     c.Endpoint,
		RedirectURL:  fmt.Sprintf("http://localhost:%d%s", c.Port, c.RedirectURI),
	}
	return config
}
