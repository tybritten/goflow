package utils

import (
	"net/http"
	"net/http/httputil"
	"time"

	log "github.com/sirupsen/logrus"
)

var httpHeaderUserAgent = "User-Agent"

func init() {
	Validator.RegisterAlias("http_method", "eq=GET|eq=HEAD|eq=POST|eq=PUT|eq=PATCH|eq=DELETE")
}

// HTTPClient is a client for HTTP requests
type HTTPClient struct {
	client           *http.Client
	defaultUserAgent string
}

// NewHTTPClient creates a new HTTP client with our default options
func NewHTTPClient(defaultUserAgent string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
			},
			Timeout: time.Duration(15 * time.Second),
		},
		defaultUserAgent: defaultUserAgent,
	}
}

func (c *HTTPClient) prepareRequest(request *http.Request) {
	// if user-agent isn't set, use our default
	if request.Header.Get(httpHeaderUserAgent) == "" {
		request.Header.Set(httpHeaderUserAgent, c.defaultUserAgent)
	}
}

// Do does the given HTTP request
func (c *HTTPClient) Do(request *http.Request) (*http.Response, error) {
	c.prepareRequest(request)

	return c.client.Do(request)
}

// DoWithDump does the given hTTP request and returns a dump of the entire request
func (c *HTTPClient) DoWithDump(request *http.Request) (*http.Response, string, error) {
	c.prepareRequest(request)

	dump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		return nil, "", err
	}

	log.Debug(string(dump))

	response, err := c.client.Do(request)

	return response, string(dump), err
}
