package webhooks

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"

	"github.com/nyaruka/goflow/flows"
)

type mockService struct {
	statusCode int
	body       string
}

// NewMockService creates a new mock HTTP service
func NewMockService(statusCode int, body string) flows.HTTPService {
	return &mockService{
		statusCode: statusCode,
		body:       body,
	}
}

func (s *mockService) Request(request *http.Request, resthook string) (flows.HTTPCall, error) {
	dump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		return nil, err
	}

	recorder := httptest.NewRecorder()
	recorder.WriteString(s.body)
	recorder.Code = s.statusCode

	response := recorder.Result()
	response.Request = request

	responseTrace, err := httputil.DumpResponse(response, false)
	if err != nil {
		return nil, err
	}

	return &call{
		url:           request.URL.String(),
		resthook:      resthook,
		request:       response.Request,
		response:      response,
		status:        StatusFromCode(response.StatusCode, resthook != ""),
		requestTrace:  string(dump),
		responseTrace: string(responseTrace),
		timeTaken:     1,
	}, nil
}

var _ flows.HTTPService = (*mockService)(nil)
