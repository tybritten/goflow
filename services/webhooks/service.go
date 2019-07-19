package webhooks

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/utils/dates"
	"github.com/pkg/errors"
)

const httpHeaderUserAgent = "User-Agent"

type service struct {
	client           *http.Client
	defaultUserAgent string
	maxBodyBytes     int
}

// NewService creates a new HTTP service
func NewService(defaultUserAgent string, maxBodyBytes int) flows.HTTPService {
	// support single tls renegotiation
	tlsConfig := &tls.Config{
		Renegotiation: tls.RenegotiateOnceAsClient,
	}

	return &service{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
				TLSClientConfig: tlsConfig,
			},
			Timeout: time.Duration(15 * time.Second),
		},
		defaultUserAgent: defaultUserAgent,
		maxBodyBytes:     maxBodyBytes,
	}
}

func (s *service) Request(request *http.Request, resthook string) (flows.HTTPCall, error) {
	// if user-agent isn't set, use our default
	if request.Header.Get(httpHeaderUserAgent) == "" {
		request.Header.Set(httpHeaderUserAgent, s.defaultUserAgent)
	}

	dump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		return nil, err
	}

	start := dates.Now()
	response, err := s.client.Do(request)
	timeTaken := dates.Now().Sub(start)

	fmt.Printf("======= %s %s =========\n%s", response.Request.Method, response.Request.URL.String(), string(dump))

	if err != nil {
		return &call{
			url:           request.URL.String(),
			request:       request,
			response:      nil,
			status:        flows.HTTPStatusConnectionError,
			requestTrace:  string(dump),
			responseTrace: err.Error(),
		}, nil
	}

	return s.newCallFromResponse(string(dump), response, s.maxBodyBytes, timeTaken, resthook)
}

// creates a new call based on the passed in http response
func (s *service) newCallFromResponse(requestTrace string, response *http.Response, maxBodyBytes int, timeTaken time.Duration, resthook string) (*call, error) {
	defer response.Body.Close()

	fmt.Printf("================= %s %s ================\n%s\n==========================================\n", response.Request.Method, response.Request.URL.String(), requestTrace)

	// save response trace without body which will be parsed separately
	responseTrace, err := httputil.DumpResponse(response, false)
	if err != nil {
		return nil, err
	}

	w := &call{
		url:           response.Request.URL.String(),
		resthook:      resthook,
		request:       response.Request,
		response:      response,
		status:        StatusFromCode(response.StatusCode, resthook != ""),
		requestTrace:  requestTrace,
		responseTrace: string(responseTrace),
		timeTaken:     timeTaken,
	}

	// we will only read up to our max body bytes limit
	bodyReader := io.LimitReader(response.Body, int64(maxBodyBytes)+1)
	var bodySniffed []byte

	// hopefully we got a content-type header
	contentTypeHeader := response.Header.Get("Content-Type")
	contentType, _, _ := mime.ParseMediaType(contentTypeHeader)

	// but if not, read first 512 bytes to sniff the content-type
	if contentType == "" {
		bodySniffed = make([]byte, 512)
		bodyBytesRead, err := bodyReader.Read(bodySniffed)
		if err != nil && err != io.EOF {
			return nil, err
		}
		bodySniffed = bodySniffed[0:bodyBytesRead]

		contentType, _, _ = mime.ParseMediaType(http.DetectContentType(bodySniffed))
	}

	// only save response body's if we have a supported content-type
	saveBody := fetchResponseContentTypes[contentType]

	if saveBody {
		bodyBytes, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			return nil, err
		}

		// if we have no remaining bytes, error because the body was too big
		if bodyReader.(*io.LimitedReader).N <= 0 {
			return nil, errors.Errorf("webhook response body exceeds %d bytes limit", maxBodyBytes)
		}

		if len(bodySniffed) > 0 {
			bodyBytes = append(bodySniffed, bodyBytes...)
		}

		w.responseTrace += string(bodyBytes)
	} else {
		w.bodyIgnored = true
	}

	return w, nil
}

var _ flows.HTTPService = (*service)(nil)
