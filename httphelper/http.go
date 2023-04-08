package httphelper

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const HeaderHost = "Host"
const HeaderXForwardedFor = "X-Forwarded-For"
const HeaderXForwardedHost = "X-Forwarded-Host"

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

func GetBaseURL(req *http.Request) string {
	protocol := getProtocolFromHostHeaders(req)
	domain := getDomainFromHostHeaders(req)
	port := getPortFromHostHeaders(req)

	return fmt.Sprintf("%s://%s%s", protocol, domain, port)
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

func getHostWithoutPort(host string) string {
	domain, _, err := net.SplitHostPort(host)
	if err != nil {
		return host
	}
	return domain
}

func getDomainFromHostHeaders(req *http.Request) string {
	domain := getHostWithoutPort(req.Header.Get(HeaderXForwardedHost))
	if domain == "" {
		domain = getHostWithoutPort(req.Host)
		if domain == "" {
			domain = "localhost"
		}
	}
	return domain
}

func getProtocolFromHostHeaders(req *http.Request) string {
	if req.Header.Get(HeaderXForwardedHost) != "" {
		return "https"
	}
	return "http"
}

func getPortFromHostHeaders(req *http.Request) string {
	// proxied requests should not have a port
	if req.Header.Get(HeaderXForwardedHost) != "" {
		return ""
	}
	_, port, _ := net.SplitHostPort(req.Host)
	if port == "80" || port == "" {
		return ""
	}
	return fmt.Sprintf(":%s", port)
}
