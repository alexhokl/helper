package googleapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"googlemaps.github.io/maps"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// NewHttpClient returns an authenticated HTTP client
// which can be used by google API clients/services
func NewHttpClient(secretFilePath string, tokenFilename string, scopes []string) (*http.Client, error) {
	fileBytes, errFile := ioutil.ReadFile(secretFilePath)
	if errFile != nil {
		return nil, errFile
	}

	oAuthConfig, errParse := google.ConfigFromJSON(fileBytes, strings.Join(scopes, " "))
	if errParse != nil {
		return nil, errParse
	}

	ctx := context.Background()
	return newAuthClient(ctx, oAuthConfig, tokenFilename)
}

func NewMapClient(apiKey string) (*maps.Client, error) {
	return maps.NewClient(maps.WithAPIKey(apiKey))
}

func newAuthClient(ctx context.Context, config *oauth2.Config, tokenFilename string) (*http.Client, error) {
	tokenFilePath, errPath := getTokenFilePath(tokenFilename)
	if errPath != nil {
		return nil, errPath
	}

	token, err := getTokenFromFile(tokenFilePath)
	if err != nil {
		token, errWeb := getTokenFromWeb(config)
		if errWeb != nil {
			return nil, errWeb
		}
		errToken := saveToken(tokenFilePath, token)
		if errToken != nil {
			return nil, errToken
		}
	}

	return config.Client(ctx, token), nil
}

func getTokenFilePath(filename string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(usr.HomeDir, url.QueryEscape(filename)), nil
}

func getTokenFromFile(path string) (*oauth2.Token, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{}
	errDecode := json.NewDecoder(file).Decode(token)
	defer file.Close()
	return token, errDecode
}

func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to [%s]\n", path)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token %v", err)
	}
	defer file.Close()
	json.NewEncoder(file).Encode(token)
	return nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code below \n%v\n", authURL)
	fmt.Printf("Authorization code: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("unable to read authorization code %v", err)
	}

	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web %v", err)
	}
	return token, nil
}
