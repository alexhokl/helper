package authhelper

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alexhokl/helper/cli"
	"golang.org/x/oauth2"
)

// BrowserOpener defines the function signature for opening URLs in a browser.
type BrowserOpener func(url string) error

// OutputWriter defines the function signature for writing user-facing output.
type OutputWriter func(w io.Writer, format string, args ...interface{}) (int, error)

// TokenSourceFactory defines the function signature for creating a token source.
type TokenSourceFactory func(ctx context.Context, config *oauth2.Config, token *oauth2.Token) oauth2.TokenSource

// TokenOptions configures the behavior of GetToken.
type TokenOptions struct {
	browserOpener   BrowserOpener
	output          OutputWriter
	outputWriter    io.Writer
	sleepDuration   time.Duration
	shutdownTimeout time.Duration
}

// TokenOption is a functional option for GetToken.
type TokenOption func(*TokenOptions)

// WithBrowserOpener sets a custom browser opener function.
func WithBrowserOpener(opener BrowserOpener) TokenOption {
	return func(o *TokenOptions) {
		o.browserOpener = opener
	}
}

// WithOutput sets a custom output writer function.
func WithOutput(writer OutputWriter) TokenOption {
	return func(o *TokenOptions) {
		o.output = writer
	}
}

// WithOutputWriter sets the io.Writer for output (default: os.Stdout).
func WithOutputWriter(w io.Writer) TokenOption {
	return func(o *TokenOptions) {
		o.outputWriter = w
	}
}

// WithSleepDuration sets the sleep duration between operations (0 to skip).
func WithSleepDuration(d time.Duration) TokenOption {
	return func(o *TokenOptions) {
		o.sleepDuration = d
	}
}

// WithShutdownTimeout sets the timeout for graceful server shutdown.
func WithShutdownTimeout(d time.Duration) TokenOption {
	return func(o *TokenOptions) {
		o.shutdownTimeout = d
	}
}

// defaultTokenOptions returns TokenOptions with production defaults.
func defaultTokenOptions() *TokenOptions {
	return &TokenOptions{
		browserOpener:   cli.OpenInBrowser,
		output:          fmt.Fprintf,
		outputWriter:    os.Stdout,
		sleepDuration:   1 * time.Second,
		shutdownTimeout: 5 * time.Second,
	}
}

// RefreshTokenOptions configures the behavior of RefreshToken.
type RefreshTokenOptions struct {
	tokenSourceFactory TokenSourceFactory
}

// RefreshTokenOption is a functional option for RefreshToken.
type RefreshTokenOption func(*RefreshTokenOptions)

// WithTokenSourceFactory sets a custom token source factory.
func WithTokenSourceFactory(factory TokenSourceFactory) RefreshTokenOption {
	return func(o *RefreshTokenOptions) {
		o.tokenSourceFactory = factory
	}
}

// defaultRefreshTokenOptions returns RefreshTokenOptions with production defaults.
func defaultRefreshTokenOptions() *RefreshTokenOptions {
	return &RefreshTokenOptions{
		tokenSourceFactory: func(ctx context.Context, config *oauth2.Config, token *oauth2.Token) oauth2.TokenSource {
			return config.TokenSource(ctx, token)
		},
	}
}
