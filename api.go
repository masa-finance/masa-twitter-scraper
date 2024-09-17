package twitterscraper

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/masa-finance/masa-twitter-scraper/httpwrap"
)

// RequestAPI get JSON from frontend API and decodes it
func (s *Scraper) RequestAPI(method, requestURL string, query url.Values, target interface{}) error {
	s.wg.Wait()
	if s.delay > 0 {
		defer s.delayRequest()
	}
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		logrus.WithError(err).Error("Invalid request URL")
		return err
	}
	csrfToken := s.getCSRFToken(parsedURL)
	var signature string
	if s.oAuthToken != "" && s.oAuthSecret != "" {
		signature = s.sign(method, parsedURL)
	}
	header := httpwrap.NewHeader()
	header.Prepare(s.userAgent, s.guestToken, s.bearerToken, csrfToken, signature, s.isLogged)
	var limitCount int
	switch method {
	case http.MethodGet:
		_, limitCount, err = s.getHTTPClient().Get(requestURL, query, header, target)
	case http.MethodPost:
		//this is only called by the logout function and will be fixed when the refactor is complete
		_, limitCount, err = s.getHTTPClient().Post(requestURL, nil, header, target)
	}
	if err != nil {
		logrus.WithError(err).Error("Failed to execute request")
		return err
	}
	if limitCount == 0 {
		s.guestToken = ""
	}
	return nil
}

func (s *Scraper) delayRequest() {
	s.wg.Add(1)
	go func() {
		time.Sleep(time.Second * time.Duration(s.delay))
		s.wg.Done()
	}()
}

func (s *Scraper) getCSRFToken(reqUrl *url.URL) string {
	for _, cookie := range s.client.GetCookies(reqUrl) {
		if cookie.Name == "ct0" {
			return cookie.Value
		}
	}
	return ""
}

// GetGuestToken from Twitter API
func (s *Scraper) GetGuestToken() error {
	header := httpwrap.NewHeader().WithBearerToken(s.bearerToken)
	if s.userAgent != "" {
		header.Add("User-Agent", s.userAgent)
	}

	result, _, err := httpwrap.NewClient().Post("https://api.twitter.com/1.1/guest/activate.json", header, nil, nil)
	if err != nil {
		return err
	}
	jsn := result.(map[string]interface{})
	var ok bool
	if s.guestToken, ok = jsn["guest_token"].(string); !ok {
		return fmt.Errorf("guest_token not found")
	}
	s.guestCreatedAt = time.Now()

	return nil
}
