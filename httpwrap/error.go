package httpwrap

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type HTTPError struct {
	Status     string
	StatusCode int
	Body       []byte
	Err        error
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("status %d: %s %v", e.StatusCode, e.Err)
}

func (e HTTPError) Log() {
	logrus.WithFields(logrus.Fields{
		"status":  e.Status,
		"content": string(e.Body),
	}).Error("Unexpected response status")
}
