package httpwrap

import "net/http"

// BearerTransport is a custom RoundTripper that adds a Bearer Token to requests.
type BearerTransport struct {
	Transport http.RoundTripper
	Token     string
}

// RoundTrip executes a single HTTP transaction and adds the Bearer Token.
func (b *BearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "Bearer "+b.Token)

	// Use the provided Transport or the default one
	if b.Transport == nil {
		b.Transport = http.DefaultTransport
	}

	return b.Transport.RoundTrip(reqClone)
}
