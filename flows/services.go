package flows

import (
	"net/http"
	"time"
)

// HTTPStatus represents the status of an HTTP call
type HTTPStatus string

const (
	// HTTPStatusSuccess represents that the webhook was successful
	HTTPStatusSuccess HTTPStatus = "success"

	// HTTPStatusConnectionError represents that the webhook had a connection error
	HTTPStatusConnectionError HTTPStatus = "connection_error"

	// HTTPStatusResponseError represents that the webhook response had a non 2xx status code
	HTTPStatusResponseError HTTPStatus = "response_error"

	// HTTPStatusSubscriberGone represents a special state of resthook responses which indicate the caller must remove that subscriber
	HTTPStatusSubscriberGone HTTPStatus = "subscriber_gone"
)

// HTTPCall is the result of an HTTP call
type HTTPCall interface {
	URL() string
	Resthook() string
	Method() string
	Status() HTTPStatus
	StatusCode() int
	TimeTaken() time.Duration
	Request() string
	Response() string
	BodyIgnored() bool
}

// HTTPService provides HTTP request making functionality to the engine
type HTTPService interface {
	Request(*http.Request, string) (HTTPCall, error)
}
