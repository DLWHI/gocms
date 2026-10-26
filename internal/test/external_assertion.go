package test

import (
	"context"
	"fmt"
)

type ExternalAssertion struct {
	Check    string   `json:"check"`
	Username string   `json:"username"`
	Password Password `json:"password"`
	Expect   string   `json:"expect"`
}

func (a *ExternalAssertion) Assert(ctx context.Context, e *Environment) error {
	var status error
	switch a.Check {
	case "login":
		actor := Actor{
			Username: a.Username,
		}
		status = actor.PerformLogin(ctx, e, a.Password)
	default:
		return fmt.Errorf("Unknown external check type: %s", a.Check)
	}
	switch a.Expect {
	case "success":
		if status != nil {
			return fmt.Errorf("Assertion failed: expected external check of %s to succeed: %v", a.Check, status)
		}
	case "fail":
		if status != nil {
			return fmt.Errorf("Assertion failed: expected external check of %s to fail", a.Check)
		}
	default:
		return fmt.Errorf("Unknown result of external check provided")
	}
	return nil
}
