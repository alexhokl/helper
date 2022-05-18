package googleapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
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

type GoogleClient struct {
	Context context.Context
	Config *oauth2.Config
	Token *oauth2.Token
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
		Config: oAuthConfig,
		Context: ctx,
		Token: &oauth2.Token{
			AccessToken: accessToken,
			RefreshToken: refreshToken,
		},
	}
	return client, nil
}

func (client *GoogleClient) GetToken() (*oauth2.Token, error) {
	client.Config.RedirectURL = fmt.Sprintf("http://localhost:%d%s", port, callbackUri)

	return getTokenFromBrowser(client, port)
}

func (client *GoogleClient) NewHttpClient() (*http.Client) {
	return client.Config.Client(client.Context, client.Token)
}

func NewMapClient(apiKey string) (*maps.Client, error) {
	return maps.NewClient(maps.WithAPIKey(apiKey))
}

func getTokenFromBrowser(client *GoogleClient, port int) (*oauth2.Token, error) {
	// add transport for self-signed certificate to context
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslcli := &http.Client{Transport: tr}
	client.Context = context.WithValue(client.Context, oauth2.HTTPClient, sslcli)

	url := client.Config.AuthCodeURL("state", oauth2.AccessTypeOffline)

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
	callbackHandler := getTokenHandler(client, tokenChannel, errorChannel)
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

	client.Token = &token

	return &token, err
}

func getTokenHandler(client *GoogleClient, tokenChannel chan oauth2.Token, errorChannel chan error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParts, _ := url.ParseQuery(r.URL.RawQuery)
		code := queryParts["code"][0]

		token, err := client.Config.Exchange(client.Context, code)
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
