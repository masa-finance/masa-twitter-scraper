package twitterscraper

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/masa-finance/masa-twitter-scraper/auth"
)

func (s *Scraper) Login(credentials ...string) error {
	var username, password, confirmation string
	if len(credentials) < 2 || len(credentials) > 3 {
		return fmt.Errorf("invalid credentials")
	}

	username, password = credentials[0], credentials[1]
	if len(credentials) == 3 {
		confirmation = credentials[2]
	}
	s.setBearerToken(BearerToken2)
	err := s.GetGuestToken()
	if err != nil {
		return err
	}

	// Create a LoginRequest using the Scraper's data
	request := auth.NewLoginRequest(s.bearerToken, s.guestToken, username, password, confirmation)

	// Call the auth.Login function
	err = auth.Login(request)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	s.isLogged = true
	s.isOpenAccount = false
	return nil
}

// LoginOpenAccount as Twitter app
func (s *Scraper) LoginOpenAccount() error {
	request := auth.NewOpenAccountRequest(s.bearerToken, s.guestToken, appConsumerKey, appConsumerSecret)

	// Call the auth.LoginOpenAccount function
	err := auth.LoginOpenAccount(request)
	if err != nil {
		return fmt.Errorf("open account failed: %w", err)
	}

	// Update the Scraper's state with the OAuth token and secret
	s.oAuthToken = request.OAuthToken
	s.oAuthSecret = request.OAuthSecret
	s.isLogged = request.IsLogged
	s.isOpenAccount = request.IsOpenAccount

	return nil
}

// Logout is reset session
func (s *Scraper) Logout() error {
	req, err := http.NewRequest("POST", auth.LogoutURL, nil)
	if err != nil {
		return err
	}
	err = s.RequestAPI(req, nil)
	if err != nil {
		return err
	}

	s.isLogged = false
	s.isOpenAccount = false
	s.guestToken = ""
	s.oAuthToken = ""
	s.oAuthSecret = ""
	s.client.WithJar()
	s.setBearerToken(BearerToken)
	return nil
}

func (s *Scraper) GetCookies() []*http.Cookie {
	var cookies []*http.Cookie
	for _, cookie := range s.client.GetCookies(twURL) {
		if strings.Contains(cookie.Name, "guest") {
			continue
		}
		cookie.Domain = twURL.Host
		cookies = append(cookies, cookie)
	}
	return cookies
}

func (s *Scraper) SetCookies(cookies []*http.Cookie) {
	s.client.SetCookies(twURL, cookies)
}

func (s *Scraper) ClearCookies() {
	s.client.WithJar()
}

func (s *Scraper) sign(method string, ref *url.URL) string {
	return auth.Sign(method, s.oAuthToken, s.oAuthSecret, appConsumerKey, appConsumerSecret, ref)
}
