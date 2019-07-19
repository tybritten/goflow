package webhooks

import (
	"net/http"
	"time"

	"github.com/nyaruka/goflow/flows"
)

// response content-types that we'll fetch
var fetchResponseContentTypes = map[string]bool{
	"application/json":       true,
	"application/javascript": true,
	"application/xml":        true,
	"text/html":              true,
	"text/plain":             true,
	"text/xml":               true,
	"text/javascript":        true,
}

// StatusFromCode determines the webhook status from the HTTP status code
func StatusFromCode(code int, isResthook bool) flows.HTTPStatus {
	// https://zapier.com/developer/documentation/v2/rest-hooks/
	if isResthook && code == 410 {
		return flows.HTTPStatusSubscriberGone
	}
	if code/100 == 2 {
		return flows.HTTPStatusSuccess
	}
	return flows.HTTPStatusResponseError
}

// a call made to an external service
type call struct {
	url           string
	resthook      string
	request       *http.Request
	response      *http.Response
	status        flows.HTTPStatus
	timeTaken     time.Duration
	requestTrace  string
	responseTrace string
	bodyIgnored   bool
}

// URL returns the full URL
func (c *call) URL() string { return c.url }

// Resthook returns the resthook slug (if this call came from a resthook action)
func (c *call) Resthook() string { return c.resthook }

// Method returns the full HTTP method
func (c *call) Method() string { return c.request.Method }

// Status returns the response status message
func (c *call) Status() flows.HTTPStatus { return c.status }

// StatusCode returns the response status code
func (c *call) StatusCode() int {
	if c.response != nil {
		return c.response.StatusCode
	}
	return 0
}

// TimeTaken returns the time taken to make the request
func (c *call) TimeTaken() time.Duration { return c.timeTaken }

// Request returns the request trace
func (c *call) Request() string { return c.requestTrace }

// Response returns the response trace
func (c *call) Response() string { return c.responseTrace }

// BodyIgnored returns whether we ignored the body because we didn't recognize the content type
func (c *call) BodyIgnored() bool {
	return c.bodyIgnored
}

var _ flows.HTTPCall = (*call)(nil)
