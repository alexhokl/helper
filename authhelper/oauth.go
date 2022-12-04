package authhelper

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/alexhokl/helper/cli"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

const lengthStateStr = 32
const lengthCodeVerifier = 64
const pkceChallengeMethod = "S256"

// character ~ is not included as some of the server implementations do not support it
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._")

type stateKey struct{}
type codeVerifierKey struct{}
type codeChallengeKey struct{}

var (
	stateContextKey         stateKey
	codeVerifierContextKey  codeVerifierKey
	codeChallengeContextKey codeChallengeKey
)

func GetToken(ctx context.Context, config *OAuthConfig, usePKCE bool) (*oauth2.Token, error) {
	if config.ClientId == "" {
		return nil, fmt.Errorf("client_id is not configured")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret is not configured")
	}
	if config.RedirectURI == "" {
		return nil, fmt.Errorf("redirect_uri is not configured")
	}

	oAuthConfig := config.GetOAuthConfig()

	// add transport for self-signed certificate to context
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslcli := &http.Client{Transport: tr}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslcli)

	state, errState := generateState(lengthStateStr)
	if errState != nil {
		return nil, errState
	}

	var codeVerifier, codeChallenge string
	if usePKCE {
		codeVerifier = generatePKCEVerifier()
		codeChallenge = generatePKCEChallenge(codeVerifier)
	}

	authOpts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
	}

	if usePKCE {
		authOpts = append(
			authOpts,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", pkceChallengeMethod),
		)
	}

	url := oAuthConfig.AuthCodeURL(state, authOpts...)

	ctx = context.WithValue(ctx, stateContextKey, state)
	ctx = context.WithValue(ctx, codeVerifierContextKey, codeVerifier)
	ctx = context.WithValue(ctx, codeChallengeContextKey, codeChallenge)

	fmt.Printf("You will now be taken to your browser for authentication [%s]\n\n", url)
	time.Sleep(1 * time.Second)
	if err := cli.OpenInBrowser(url); err != nil {
		return nil, err
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

func RefreshToken(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve a new token from refresh token: %v", err)
	}
	return newToken, nil
}

func SaveTokenToViper(token *oauth2.Token) error {
	if token == nil {
		return fmt.Errorf("token is not specified")
	}

	viper.Set("access_token", token.AccessToken)
	viper.Set("refresh_token", token.RefreshToken)
	viper.Set("token_type", token.TokenType)
	viper.Set("expiry", token.Expiry)

	viper.WriteConfig()

	return nil
}

func LoadTokenFromViper() (*oauth2.Token, error) {
	accessToken := viper.GetString("access_token")
	refreshToken := viper.GetString("refresh_token")
	tokenType := viper.GetString("token_type")
	expiry := viper.GetTime("expiry")

	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		Expiry:       expiry,
	}

	return token, nil
}

func getTokenHandler(ctx context.Context, config *oauth2.Config, tokenChannel chan oauth2.Token, errorChannel chan error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParts, _ := url.ParseQuery(r.URL.RawQuery)

		state := queryParts.Get("state")
		expectedState := ctx.Value(stateContextKey).(string)
		if strings.Compare(state, expectedState) != 0 {
			errorChannel <- fmt.Errorf("state mismatch (expected %s, got %s)", expectedState, state)
			return
		}

		codeChallenge := queryParts.Get("code_challenge")
		expectedCodeChallenge := ctx.Value(codeChallengeContextKey).(string)
		if strings.Compare(codeChallenge, expectedCodeChallenge) != 0 {
			errorChannel <- fmt.Errorf("code_challenge mismatch (expected %s, got %s)", expectedCodeChallenge, codeChallenge)
			return
		}

		if expectedCodeChallenge != "" {
			codeChallengeMethod := queryParts.Get("code_challenge_method")
			if codeChallengeMethod != pkceChallengeMethod {
				errorChannel <- fmt.Errorf("code_challenge_method mismatch (expected %s, got %s)", pkceChallengeMethod, codeChallengeMethod)
				return
			}
		}

		if queryParts.Get("error") != "" {
			if errorDescription := queryParts.Get("error_description"); errorDescription != "" {
				errorChannel <- fmt.Errorf("%s: %s", queryParts.Get("error"), errorDescription)
				return
			}
			errorChannel <- fmt.Errorf("%s", queryParts.Get("error"))
			return
		}

		authOpts := []oauth2.AuthCodeOption{}

		if expectedCodeChallenge != "" {
			authOpts = append(
				authOpts,
				oauth2.SetAuthURLParam("code_verifier", ctx.Value(codeVerifierContextKey).(string)),
				oauth2.SetAuthURLParam("code_challenge_method", pkceChallengeMethod),
			)
		}

		token, err := config.Exchange(
			ctx,
			queryParts.Get("code"),
			authOpts...,
		)
		if err != nil {
			errorChannel <- fmt.Errorf("failed to exchange token: %v", err)
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

func generateState(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func generatePKCEVerifier() string {
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, lengthCodeVerifier)
	for i := range b {
		b[i] = letterRunes[generator.Intn(len(letterRunes))]
	}
	return string(b)
}

func generatePKCEChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	// see https://www.rfc-editor.org/rfc/rfc7636#appendix-A
	// for no-padding requirement
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
}
