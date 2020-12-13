package http

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// SetBearerTokenHeader sets the Authorization header with the specified bearer
// token
func SetBearerTokenHeader(req *http.Request, token string) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
}

// DumpResponse dumps the body of the specified response to stdout
func DumpResponse(resp *http.Response) error {
	_, err := io.Copy(os.Stdout, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
