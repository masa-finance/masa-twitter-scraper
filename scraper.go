package twitterscraper

import (
	"net/http"
	"sync"
	"time"

	"github.com/masa-finance/masa-twitter-scraper/auth"
	"github.com/masa-finance/masa-twitter-scraper/httpwrap"
	"github.com/masa-finance/masa-twitter-scraper/types"
)

// Scraper object
type Scraper struct {
	bearerToken    string
	client         *httpwrap.Client
	delay          int64
	guestToken     string
	guestCreatedAt time.Time
	includeReplies bool
	isLogged       bool
	isOpenAccount  bool
	oAuthToken     string
	oAuthSecret    string
	proxy          string
	searchMode     SearchMode
	wg             sync.WaitGroup
	userAgent      string
}

// SearchMode type
type SearchMode int

const (
	// SearchTop - default mode
	SearchTop SearchMode = iota
	// SearchLatest - live mode
	SearchLatest
	// SearchPhotos - image mode
	SearchPhotos
	// SearchVideos - video mode
	SearchVideos
	// SearchUsers - user mode
	SearchUsers
)

// SetUserAgent sets the user agent for the scraper
func (s *Scraper) SetUserAgent(userAgent string) *Scraper {
	s.userAgent = userAgent
	return s
}

// getHTTPClient returns the configured http.Client
func (s *Scraper) getHTTPClient() *httpwrap.Client {
	return s.client
}

// New creates a Scraper object
func New() *Scraper {
	scraper := &Scraper{
		bearerToken: BearerToken,
		client:      httpwrap.NewClient().WithJar().WithBearerToken(BearerToken),
	}
	scraper.SetUserAgent(auth.GetRandomUserAgent())
	return scraper
}

func (s *Scraper) GetBearerToken() string {
	return s.bearerToken
}

func (s *Scraper) setBearerToken(token string) {
	s.bearerToken = token
	s.guestToken = ""
}

// IsGuestToken check if guest token not empty
func (s *Scraper) IsGuestToken() bool {
	return s.guestToken != ""
}

// SetSearchMode switcher
func (s *Scraper) SetSearchMode(mode SearchMode) *Scraper {
	s.searchMode = mode
	return s
}

// WithDelay add delay between API requests (in seconds)
func (s *Scraper) WithDelay(seconds int64) *Scraper {
	s.delay = seconds
	return s
}

// WithReplies enable/disable load timeline with tweet replies
func (s *Scraper) WithReplies(b bool) *Scraper {
	s.includeReplies = b
	return s
}

// client timeout
func (s *Scraper) WithClientTimeout(timeout time.Duration) *Scraper {
	s.client.SetTimeout(timeout)
	return s
}

// SetProxy
// set http proxy in the format `http://HOST:PORT`
// set socket proxy in the format `socks5://HOST:PORT`
func (s *Scraper) SetProxy(proxyAddr string) error {
	return s.client.SetProxy(proxyAddr)
}

// IsLoggedIn check if scraper logged in
func (s *Scraper) IsLoggedIn() bool {
	s.isLogged = true
	s.setBearerToken(BearerToken2)
	req, err := http.NewRequest("GET", VerifyCredentialsURL, nil)
	if err != nil {
		return false
	}
	var verify types.VerifyCredentials
	err = s.RequestAPI(req, &verify)
	if err != nil || verify.Errors != nil {
		s.isLogged = false
		s.setBearerToken(BearerToken)
	} else {
		s.isLogged = true
	}
	return s.isLogged
}
