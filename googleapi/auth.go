package googleapi

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	"googlemaps.github.io/maps"

	"github.com/alexhokl/helper/authhelper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const port = 9998
const callbackUri = "/callback"

type GoogleClient struct {
	Context context.Context
	Config  *oauth2.Config
	Token   *oauth2.Token
}

func New(ctx context.Context, googleClientSecretFilePath string, accessToken string, refreshToken string, scopes []string) (*GoogleClient, error) {
	fileBytes, errFile := ioutil.ReadFile(googleClientSecretFilePath)
	if errFile != nil {
		return nil, errFile
	}

	oAuthConfig, errParse := google.ConfigFromJSON(fileBytes, strings.Join(scopes, " "))
	if errParse != nil {
		return nil, errParse
	}
	client := &GoogleClient{
		Config:  oAuthConfig,
		Context: ctx,
		Token: &oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
	return client, nil
}

func (client *GoogleClient) GetToken() (*oauth2.Token, error) {
	token, err := authhelper.GetToken(
		client.Context,
		authhelper.OAuthConfig{
			ClientId:     client.Config.ClientID,
			ClientSecret: client.Config.ClientSecret,
			Scopes:       client.Config.Scopes,
			RedirectURI:  callbackUri,
			Port:         port,
			Endpoint:     google.Endpoint,
		},
	)
	if err != nil {
		return nil, err
	}
	client.Token = token

	return token, nil
}

func (client *GoogleClient) NewHttpClient() *http.Client {
	return client.Config.Client(client.Context, client.Token)
}

func NewMapClient(apiKey string) (*maps.Client, error) {
	return maps.NewClient(maps.WithAPIKey(apiKey))
}
