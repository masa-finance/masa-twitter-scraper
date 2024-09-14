package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/url"
	"strconv"
	"time"
)

// Sign generates an OAuth 1.0a signature for an HTTP request.
// It constructs the OAuth signature base string and uses HMAC-SHA1 to sign it,
// returning the complete Authorization header value.
//
// Parameters:
// - httpMethod: The HTTP method (e.g., "GET", "POST") used for the request.
// - oauthToken: The OAuth token provided by the service for authentication.
// - oauthTokenSecret: The secret associated with the OAuth token.
// - consumerKey: The consumer key provided by the service for your application.
// - consumerSecret: The secret associated with the consumer key.
// - requestURL: The URL of the request, including any query parameters.
//
// Returns:
// A string representing the Authorization header value, which includes the
// OAuth parameters and the generated signature.
//
// Note:
//   - The function generates a unique nonce and timestamp for each call to ensure
//     the signature's uniqueness and prevent replay attacks.
//   - Ensure that the consumer secret and token secret are kept secure, as they
//     are critical for generating valid signatures.
func Sign(httpMethod, oauthToken, oauthTokenSecret, consumerKey, consumerSecret string, requestURL *url.URL) string {
	oauthParams := make(map[string]string)
	oauthParams["oauth_consumer_key"] = consumerKey
	oauthParams["oauth_nonce"] = strconv.FormatInt(time.Now().UnixNano(), 10) // Use a unique nonce
	oauthParams["oauth_signature_method"] = "HMAC-SHA1"
	oauthParams["oauth_timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	oauthParams["oauth_token"] = oauthToken

	signingKey := []byte(consumerSecret + "&" + oauthTokenSecret)
	hmacHasher := hmac.New(sha1.New, signingKey)

	queryParams := requestURL.Query()
	for key, value := range oauthParams {
		queryParams.Set(key, value)
	}

	signatureBaseComponents := []string{httpMethod, requestURL.Scheme + "://" + requestURL.Host + requestURL.Path, queryParams.Encode()}
	var signatureBaseBuffer bytes.Buffer
	for _, component := range signatureBaseComponents {
		if signatureBaseBuffer.Len() > 0 {
			signatureBaseBuffer.WriteByte('&')
		}
		signatureBaseBuffer.WriteString(url.QueryEscape(component))
	}
	hmacHasher.Write(signatureBaseBuffer.Bytes())

	oauthParams["oauth_signature"] = base64.StdEncoding.EncodeToString(hmacHasher.Sum(nil))

	var authorizationHeaderBuffer bytes.Buffer
	for key, value := range oauthParams {
		if authorizationHeaderBuffer.Len() > 0 {
			authorizationHeaderBuffer.WriteByte(',')
		}
		authorizationHeaderBuffer.WriteString(key)
		authorizationHeaderBuffer.WriteByte('=')
		authorizationHeaderBuffer.WriteString(url.QueryEscape(value))
	}

	return "OAuth " + authorizationHeaderBuffer.String()
}
