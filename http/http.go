package http

import (
	"fmt"
	"net/http"
)

// SetBearerTokenHeader sets the Authorization header with the specified bearer
// token
func SetBearerTokenHeader(req *http.Request, token string) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
}
