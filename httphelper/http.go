package httphelper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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

func HasContentType(resp *http.Response, contentType string) bool {
	contentTypeHeader := resp.Header.Get("Content-Type")
	return hasContentType(contentTypeHeader, contentType)
}

func IsSuccessResponse(resp *http.Response) bool {
	return resp.StatusCode < 300
}

func GetHeaders(keyValues map[string]string) http.Header {
	headers := http.Header{}
	for key, value := range keyValues {
		headers.Set(key, value)
	}
	return headers
}

func GetURL(path string, params map[string]string) (*url.URL, error) {
	parsedURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	for k, v := range params {
		queryParams.Set(k, v)
	}

	query := parsedURL.Query()
	for k, v := range queryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}
	parsedURL.RawQuery = query.Encode()
	return parsedURL, nil
}

func NewRequest(method string, path string, params map[string]string, headers map[string]string, body io.Reader) (*http.Request, error) {
	parsedURL, err := GetURL(path, params)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(method, parsedURL.String(), body)
	if err != nil {
		return nil, err
	}

	request.Header = GetHeaders(headers)
	return request, nil
}

func hasContentType(header string, contentType string) bool {
	contentTypes := strings.Split(header, ";")
	for _, t := range contentTypes {
		if strings.TrimSpace(t) == contentType {
			return true
		}
	}
	return false
}
