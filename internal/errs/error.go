package errs

import (
	"errors"
	"fmt"
	"net/http"
)

type Constant string

func (e Constant) Error() string {
	return string(e)
}

const (
	ErrNotFound = Constant("Entity not found")
	ErrInval    = Constant("Invalid argument")
	ErrCred     = Constant("Invalid username or password")
	ErrExpired  = Constant("Session expired")
	ErrPerm     = Constant("Insufficient permissions")
	ErrLock     = Constant("Record is locked")
)

func ServeMessage(w http.ResponseWriter, entity string, err error) {
	if errors.Is(err, ErrCred) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	} else if errors.Is(err, ErrPerm) {
		http.Error(w, "Insufficent permissions", http.StatusUnauthorized)
	} else if errors.Is(err, ErrExpired) {
		http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
	} else if errors.Is(err, ErrInval) {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
	} else if errors.Is(err, ErrNotFound) {
		http.Error(w, fmt.Sprintf("%s not found", entity), http.StatusNotFound)
	} else {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
