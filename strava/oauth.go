package strava

import (
	"fmt"

	"golang.org/x/oauth2"
)

const apiURL = "https://www.strava.com"

// GetOAuthEndpoint returns the OAuth endpoint for Strava
func GetOAuthEndpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/authorize", apiURL),
			TokenURL: fmt.Sprintf("%s/oauth/token", apiURL),
	}
}
