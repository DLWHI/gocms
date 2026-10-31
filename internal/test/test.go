package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Result struct {
	ResponseStatus int
	Message        string
	Reason         error
}

type Case struct {
	ID     string   `json:"id"`
	Group  string   `json:"group"`
	Title  string   `json:"title"`
	Source string   `json:"source"`
	Tags   []string `json:"tags"`
	Steps  []Step   `json:"steps"`
}

type Step struct {
	Actor      string         `json:"actor"`
	Request    HttpRequest    `json:"request"`
	Expect     ExpectedResult `json:"expect"`
	Eventually uint           `json:"eventually_ms"`
}

func (s *Step) Perform(ctx context.Context, d *Data, e *Environment) *Result {
	actor, ok := d.Fixtures.Actors[s.Actor]
	if !ok && s.Actor != "anonymous" {
		return s.fail(nil, fmt.Errorf("Requested step for unknown actor"))
	}

	res, err := s.Request.Perform(ctx, e, actor)
	if err != nil {
		return s.fail(res, err)
	}
	defer res.Body.Close()

	if err := e.Outbox.Dispatch(ctx, 5); err != nil {
		return s.fail(res, err)
	}

	if err := s.Expect.Check(d, res); err != nil {
		return s.fail(res, err)
	}

	for _, external := range s.Expect.ExternalChecks {
		if err := external.Assert(ctx, e); err != nil {
			return s.fail(res, err)
		}
	}
	return nil
}

func (s *Step) fail(res *http.Response, reason error) *Result {
	result := &Result{Reason: reason}
	if res != nil {
		result.ResponseStatus = res.StatusCode
		data, err := io.ReadAll(res.Body)
		if err != nil {
			result.Message = "Failed to read failed response body"
		}
		result.Message = string(data)
	}
	return result
}
