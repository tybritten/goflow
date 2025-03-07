package test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
)

// NewTestHTTPServer sets up a mock server for webhook actions
func NewTestHTTPServer(port int) *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(testHTTPHandler))

	if port > 0 {
		// manually create a listener for our test server so that our output is predictable
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			panic(err.Error())
		}
		server.Listener = l
	}
	server.Start()
	return server
}

func testHTTPHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	statusCode := http.StatusOK
	contentType := r.URL.Query().Get("type")
	data := []byte(r.URL.Query().Get("content"))

	cmd := r.URL.Query().Get("cmd")
	switch cmd {
	case "success":
		contentType = "text/plain; charset=utf-8"
		data = []byte(`{ "ok": "true" }`)
	case "binary":
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		sizeParam := r.URL.Query().Get("size")
		if sizeParam == "" {
			sizeParam = "10"
		}
		size, _ := strconv.Atoi(sizeParam)
		data = make([]byte, size)
		for i := 0; i < size; i++ {
			data[i] = byte(40 + i%10)
		}

		w.Header().Set("Content-Length", sizeParam)
	case "textjs":
		contentType = "text/javascript; charset=iso-8859-1"
		data = []byte(`{ "ok": "true" }`)
	case "typeless":
		w.Header().Set("Content-Type", "")
	case "unavailable":
		statusCode = http.StatusServiceUnavailable
		data = []byte(`{ "errors": ["service unavailable"] }`)
	case "badrequest":
		statusCode = http.StatusBadRequest
		data = []byte(`{ "errors": ["bad_request"] }`)
	case "gone":
		statusCode = http.StatusGone
		data = []byte(`{ "errors": ["gone"] }`)
	}

	w.Header().Set("Date", "Wed, 11 Apr 2018 18:24:30 GMT")
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	w.WriteHeader(statusCode)
	w.Write(data)
}
