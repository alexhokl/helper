package googleapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"googlemaps.github.io/maps"

	"github.com/alexhokl/helper/cli"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const port = 9998
const callbackUri = "/callback"

func GetToken(ctx context.Context, googleClientSecretFilePath string, scopes []string) (*oauth2.Token, error) {
	fileBytes, errFile := ioutil.ReadFile(googleClientSecretFilePath)
	if errFile != nil {
		return nil, errFile
	}

	oAuthConfig, errParse := google.ConfigFromJSON(fileBytes, strings.Join(scopes, " "))
	if errParse != nil {
		return nil, errParse
	}
	oAuthConfig.RedirectURL = fmt.Sprintf("http://localhost:%d%s", port, callbackUri)

	return getTokenFromBrowser(ctx, oAuthConfig, port)
}

func NewHttpClient(ctx context.Context, googleClientSecretFilePath string, token *oauth2.Token) (*http.Client, error) {
	fileBytes, errFile := ioutil.ReadFile(googleClientSecretFilePath)
	if errFile != nil {
		return nil, errFile
	}

	oAuthConfig, errParse := google.ConfigFromJSON(fileBytes)
	if errParse != nil {
		return nil, errParse
	}
	return oAuthConfig.Client(ctx, token), nil
}

// NewHttpClientAndSaveToken returns an authenticated HTTP client
// which can be used by google API clients/services
func NewHttpClientAndSaveToken(secretFilePath string, tokenFilename string, scopes []string) (*http.Client, error) {
	fileBytes, errFile := ioutil.ReadFile(secretFilePath)
	if errFile != nil {
		return nil, errFile
	}

	oAuthConfig, errParse := google.ConfigFromJSON(fileBytes, strings.Join(scopes, " "))
	if errParse != nil {
		return nil, errParse
	}
	oAuthConfig.RedirectURL = fmt.Sprintf("http://localhost:%d%s", port, callbackUri)

	ctx := context.Background()
	token, err := getTokenAndSave(ctx, oAuthConfig, tokenFilename, port)
	if err != nil {
		return nil, err
	}
	return oAuthConfig.Client(ctx, token), nil
}

func NewMapClient(apiKey string) (*maps.Client, error) {
	return maps.NewClient(maps.WithAPIKey(apiKey))
}

func getTokenAndSave(ctx context.Context, config *oauth2.Config, tokenFilename string, port int) (*oauth2.Token, error) {
	tokenFilePath, errPath := getTokenFilePath(tokenFilename)
	if errPath != nil {
		return nil, errPath
	}

	token, err := getTokenFromFile(tokenFilePath)
	if err == nil {
		return token, nil
	}
	token, errWeb := getTokenFromBrowser(ctx, config, port)
	if errWeb != nil {
		return nil, errWeb
	}
	return token, nil
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

func getTokenFromBrowser(ctx context.Context, config *oauth2.Config, port int) (*oauth2.Token, error) {
	// add transport for self-signed certificate to context
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslcli := &http.Client{Transport: tr}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslcli)

	url := config.AuthCodeURL("state", oauth2.AccessTypeOffline)

	fmt.Println("You will now be taken to your browser for authentication")
	time.Sleep(1 * time.Second)
	cmdName, cmdArgs := cli.GetOpenCommand(url)
	_, errOpen := exec.Command(cmdName, cmdArgs...).Output()
	if errOpen != nil {
		return nil, errOpen
	}
	time.Sleep(1 * time.Second)

	var token oauth2.Token
	var err error
	tokenDone := &sync.WaitGroup{}
	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)
	serverErrorChannel := make(chan error, 1)
	callbackHandler := getTokenHandler(ctx, config, tokenChannel, errorChannel)
	http.HandleFunc(callbackUri, callbackHandler)
	server := getServer(port)
	go func(tokens chan oauth2.Token) {
		for t := range tokens {
			token = t
			tokenDone.Done()
			return
		}
	}(tokenChannel)
	go func(errs chan error) {
		for e := range errs {
			err = e
			tokenDone.Done()
			return
		}
	}(errorChannel)
	tokenDone.Add(1)
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			serverErrorChannel <- err
		}
		tokenDone.Done()
	}()
	tokenDone.Wait()

	return &token, err
}

func getTokenHandler(ctx context.Context, config *oauth2.Config, tokenChannel chan oauth2.Token, errorChannel chan error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParts, _ := url.ParseQuery(r.URL.RawQuery)
		code := queryParts["code"][0]

		token, err := config.Exchange(ctx, code)
		if err != nil {
			errorChannel <- err
			return
		}

		tokenChannel <- *token
		fmt.Fprintf(w, "You have been authenticated. This browser window can be closed.")
	}
}

func getServer(port int) *http.Server {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}
	return server
}
