package airtable

import (
	"fmt"

	"golang.org/x/oauth2"
)

const apiURL = "https://airtable.com"

// GetOAuthEndpoint returns the OAuth endpoint for Airtable
func GetOAuthEndpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s/oauth2/v1/authorize", apiURL),
		TokenURL: fmt.Sprintf("%s/oauth2/v1/token", apiURL),
	}
}
