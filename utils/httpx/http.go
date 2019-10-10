package httpx

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/nyaruka/goflow/utils/dates"
)

var debug = false

// Do makes the given HTTP request using the current requestor
func Do(client *http.Client, request *http.Request) (*http.Response, error) {
	return currentRequestor.Do(client, request)
}

// Trace holds the complete trace of an HTTP request/response
type Trace struct {
	Request       *http.Request
	Response      *http.Response
	Body          []byte
	RequestTrace  []byte
	ResponseTrace []byte
	TimeTaken     time.Duration
}

func (t *Trace) String() string {
	b := &strings.Builder{}
	b.WriteString(fmt.Sprintf(">>>>>>>> %s %s\n", t.Request.Method, t.Request.URL))
	b.WriteString(string(t.RequestTrace))
	b.WriteString("\n<<<<<<<<\n")
	b.WriteString(string(t.ResponseTrace))
	return b.String()
}

// DoTrace makes the given request saving traces of the complete request and response
func DoTrace(client *http.Client, method string, url string, body io.Reader, headers map[string]string) (*Trace, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	requestTrace, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		return nil, err
	}

	start := dates.Now()
	response, err := Do(client, request)
	timeTaken := dates.Now().Sub(start)

	if err != nil {
		return nil, err
	}

	// save response trace without body which will be parsed separately
	responseTrace, err := httputil.DumpResponse(response, false)
	if err != nil {
		return nil, err
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// add read body to response trace
	responseTrace = append(responseTrace, responseBody...)
	trace := &Trace{
		Request:       request,
		Response:      response,
		RequestTrace:  requestTrace,
		ResponseTrace: responseTrace,
		Body:          responseBody,
		TimeTaken:     timeTaken,
	}

	if debug {
		fmt.Println(trace.String())
	}

	return trace, nil
}

// Requestor is anything that can make an HTTP request with a client
type Requestor interface {
	Do(*http.Client, *http.Request) (*http.Response, error)
}

type defaultRequestor struct{}

func (r defaultRequestor) Do(client *http.Client, request *http.Request) (*http.Response, error) {
	return client.Do(request)
}

// DefaultRequestor is the default HTTP requestor
var DefaultRequestor Requestor = defaultRequestor{}
var currentRequestor = DefaultRequestor

// SetRequestor sets the requestor used by Request
func SetRequestor(requestor Requestor) {
	currentRequestor = requestor
}

func SetDebug(enabled bool) {
	debug = enabled
}
