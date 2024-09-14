package twitterscraper

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock HTTP server to simulate Twitter's API response
func mockTwitterServer() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/1.1/guest/activate.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"guest_token": "mocked_guest_token"}`))
		if err != nil {
			return
		}
	})
	return httptest.NewServer(handler)
}

func TestGetGuestToken(t *testing.T) {
	// Create a mock server
	server := mockTwitterServer()
	defer server.Close()

	//// Replace the Twitter API URL with the mock server URL
	//originalURL := "https://api.twitter.com/1.1/guest/activate.json"
	//mockURL := server.URL + "/1.1/guest/activate.json"

	// Create a new Scraper instance
	scraper := New()

	// Mock the HTTP client or method to use the mockURL
	// This might involve setting a field in the Scraper or modifying the GetGuestToken method to accept a URL

	// Call the GetGuestToken method
	err := scraper.GetGuestToken()

	// Check for errors
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the guest token is set correctly
	expectedToken := "mocked_guest_token"
	if scraper.guestToken != expectedToken {
		t.Errorf("expected guest token %v, got %v", expectedToken, scraper.guestToken)
	}
}
