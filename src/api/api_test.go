package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func makeRequest[REQ, RES any](h http.Handler, method, url string, body *REQ) (int, RES, error) {
	var res RES

	// marshal request body
	var rbody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, res, fmt.Errorf("marshal body: %w", err)
		}
		rbody = bytes.NewReader(data)
	}

	// create request
	req, err := http.NewRequest(method, url, rbody)
	if err != nil {
		return 0, res, fmt.Errorf("new request: %w", err)
	}

	// do request
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// unmarshal response body
	if rec.Body != nil && rec.Body.Len() > 0 {
		if err := json.Unmarshal((*rec.Body).Bytes(), &res); err != nil {
			return 0, res, fmt.Errorf("unmarshal body: %w", err)
		}
	}

	return rec.Code, res, nil
}
