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
	Length uint             `json:"json_padded_bytes"`
	Base   map[string]Value `json:"base"`
}

func (e *Environment) PerformRequest(ctx context.Context, actor *Actor, request HttpRequest) (*http.Response, error) {
	url, err := RequestURL(request.Path, request.PathParams, e.TestData)
	if err != nil {
		return nil, err
	}
	body, err := BodyReaderOf(request, e.TestData)
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequestWithContext(ctx, request.Method, url, body)

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
		return request.BodyGenerator.Create(testdata)
	case request.Body != nil:
		payload, err := testdata.ResolveMap(request.Body)
		if err != nil {
			return nil, err
		}
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(&payload); err != nil {
			return nil, err
		}
		return io.NopCloser(buf), nil
	default:
		return nil, nil
	}
}

func (g *JSONGenerator) Create(testdata *Data) (io.ReadCloser, error) {
	payload, err := testdata.ResolveMap(g.Base)
	if err != nil {
		return nil, err
	}
	payload["_"] = ""

	base, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if uint(len(base)) > g.Length {
		return nil, fmt.Errorf("base is %d bytes, larger than requested %d", len(base), g.Length)
	}
	payload["_"] = strings.Repeat(" ", int(g.Length)-len(base))

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if len(body) != int(g.Length) {
		return nil, fmt.Errorf("generated %d bytes, want %d", len(body), g.Length)
	}
	return io.NopCloser(bytes.NewReader(body)), nil
}
