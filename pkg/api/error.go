package api

import (
	"fmt"
	"net/http"
)

var (
	contentTypeHeader = http.CanonicalHeaderKey("content-type")
)

const (
	badRequest    = "Bad Request: %s"
	internalError = "Internal Error"
	lockedError   = "Resource Locked"
)

// BadRequest handles returning a simple reason for a bad request HTTP error.
func BadRequest(res http.ResponseWriter, reason string) {
	// We should ensure the proper content-type header here.
	res.Header().Set(contentTypeHeader, "text/plain")

	// Write the actual status.
	res.WriteHeader(http.StatusBadRequest)

	// Write the error message.
	if _, err := res.Write([]byte(fmt.Sprintf(badRequest, reason))); err != nil {
		fmt.Println("Failed to write bad request error to client.", err)
	}
}

// InternalError handles returning a internal server error.
func InternalError(res http.ResponseWriter) {
	// We should ensure the proper content-type header here.
	res.Header().Set(contentTypeHeader, "text/plain")

	// Write the actual status.
	res.WriteHeader(http.StatusInternalServerError)

	// Write the error message.
	if _, err := res.Write([]byte(internalError)); err != nil {
		fmt.Println("Failed to write internal error to client.", err)
	}
}

// Locked handles returning a locked error.
func Locked(res http.ResponseWriter) {
	// We should ensure the proper content-type header here.
	res.Header().Set(contentTypeHeader, "text/plain")

	// Write the actual status.
	res.WriteHeader(http.StatusLocked)

	// Write the error message.
	if _, err := res.Write([]byte(lockedError)); err != nil {
		fmt.Println("Failed to write internal error to client.", err)
	}
}
