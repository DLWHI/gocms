package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

type HttpRequest struct {
	Method        string           `json:"method"`
	Path          string           `json:"path"`
	PathParams    map[string]Value `json:"path_params"`
	Body          map[string]Value `json:"body"`
	BodyGenerator *JSONGenerator   `json:"body_generator"`
	BodyRaw       *string          `json:"body_raw"`
	Headers       map[string]Value `json:"headers"`
}

type JSONGenerator struct {
	Padding uint             `json:"json_padded_bytes"`
	Base    map[string]Value `json:"base"`
}

func (e *Environment) PerformRequest(ctx context.Context, actor *Actor, request HttpRequest) (*http.Response, error) {
	url, err := RequestURL(request.Path, request.PathParams, e.TestData)
	if err != nil {
		return nil, err
	}
	req := httptest.NewRequestWithContext(ctx, request.Method, url, nil)

	req.Header.Set("User-Agent", "Test-Client")

	if actor != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", actor.Session))
	}

	for key, value := range request.Headers {
		val, err := e.TestData.Resolve(value)
		if err != nil {
			return nil, err
		}
		req.Header.Set(key, val)
	}

	req.Body, err = BodyReaderOf(request, e.TestData)
	if err != nil {
		return nil, err
	}

	w := httptest.NewRecorder()

	e.Router.ServeHTTP(w, req)
	return w.Result(), nil
}

func RequestURL(pathname string, params map[string]Value, testdata *Data) (string, error) {
	for key, val := range params {
		v, err := testdata.Resolve(val)
		if err != nil {
			return "", err
		}
		pathname = strings.ReplaceAll(pathname, "{"+key+"}", v)
	}
	return pathname, nil
}

func BodyReaderOf(request HttpRequest, testdata *Data) (io.ReadCloser, error) {
	switch {
	case request.BodyRaw != nil:
		return io.NopCloser(strings.NewReader(*request.BodyRaw)), nil
	case request.BodyGenerator != nil:
		buf := &bytes.Buffer{}
		val, err := testdata.Resolve(request.BodyGenerator.Base)
		if err != nil {
			return nil, err
		}
		buf.Write([]byte(val))
		buf.Write(make([]byte, request.BodyGenerator.Padding))
		return io.NopCloser(buf), nil
	case request.Body != nil:
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(request.Body); err != nil {
			return nil, err
		}
		return io.NopCloser(buf), nil
	default:
		return nil, nil
	}
}
