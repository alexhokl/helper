package googleapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func New(ctx context.Context, googleClientSecretFilePath string, accessToken *oauth2.Token, scopes []string) (*GoogleClient, error) {
	path := filepath.Clean(googleClientSecretFilePath)
	file, errOpen := os.Open(path)
	if errOpen != nil {
		return nil, fmt.Errorf("unable to open file %s with error %w", path, errOpen)
	}
	defer func() {
		_ = file.Close()
	}()

	fileBytes, errFile := io.ReadAll(file)
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
		Token:   accessToken,
	}
	return client, nil
}

func NewFromClientIDSecret(ctx context.Context, clientID string, clientSecret string, scopes []string, accessToken *oauth2.Token) (*GoogleClient, error) {
	client := &GoogleClient{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
			RedirectURL:  fmt.Sprintf("http://localhost:%d%s", port, callbackUri),
		},
		Context: ctx,
		Token:   accessToken,
	}
	return client, nil
}

func (client *GoogleClient) GetToken() (*oauth2.Token, error) {
	config := &authhelper.OAuthConfig{
		ClientId:     client.Config.ClientID,
		ClientSecret: client.Config.ClientSecret,
		Scopes:       client.Config.Scopes,
		RedirectURI:  callbackUri,
		Port:         port,
		Endpoint:     google.Endpoint,
	}
	token, err := authhelper.GetToken(client.Context, config, false)
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
