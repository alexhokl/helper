package authhelper

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/alexhokl/helper/cli"
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

func GetToken(ctx context.Context, config OAuthConfig) (*oauth2.Token, error) {
	if config.ClientId == "" {
		return nil, fmt.Errorf("client_id is not configured")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret is not configured")
	}
	if config.RedirectURI == "" {
		return nil, fmt.Errorf("redirect_uri is not configured")
	}

	oAuthConfig := &oauth2.Config{
		ClientID:     config.ClientId,
		ClientSecret: config.ClientSecret,
		Scopes:       config.Scopes,
		Endpoint:     config.Endpoint,
		RedirectURL:  fmt.Sprintf("http://localhost:%d%s", config.Port, config.RedirectURI),
	}

	// add transport for self-signed certificate to context
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslcli := &http.Client{Transport: tr}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslcli)

	url := oAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)

	fmt.Println("You will now be taken to your browser for authentication")
	time.Sleep(1 * time.Second)
	cmdName, cmdArgs := cli.GetOpenCommand(url)
	_, errOpen := exec.Command(cmdName, cmdArgs...).Output()
	if errOpen != nil {
		cmdParts := []string { cmdName }
		cmdParts = append(cmdParts, cmdArgs...)
		return nil, fmt.Errorf("unable to complete command [%s] %w", strings.Join(cmdParts, " "), errOpen)
	}
	time.Sleep(1 * time.Second)

	var token oauth2.Token
	var err error
	tokenDone := &sync.WaitGroup{}
	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)
	serverErrorChannel := make(chan error, 1)
	callbackHandler := getTokenHandler(ctx, oAuthConfig, tokenChannel, errorChannel)
	http.HandleFunc(config.RedirectURI, callbackHandler)
	server := getServer(config.Port)
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
