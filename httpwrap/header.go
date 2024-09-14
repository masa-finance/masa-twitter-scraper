package httpwrap

import (
	"encoding/base64"
)

// A Header represents the key-value pairs in an HTTP header.
// It wraps the Go standard library's httpwrap.Header type.
// It is not an array of strings, so it won't work if you have multiple headers with the same key and order matters.
type Header map[string]string

func NewHeader() Header {
	return Header{}
}

// AddBasicAuth adds an Authorization header with Basic Authentication.
func (h Header) AddBasicAuth(username, password string) {
	base64Value := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	h["Authorization"] = "Basic " + base64Value
}

// AddContentType add the content type to the header.
func (h Header) AddContentType(contentType string) {
	h["Content-Type"] = contentType
}

// Add adds a key-value pair to the header.
func (h Header) Add(key, value string) {
	h[key] = value
}

func (h Header) WithBearerToken(token string) Header {
	h["Authorization"] = "Bearer " + token
	return h
}
