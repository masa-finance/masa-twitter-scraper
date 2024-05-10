package twitterscraper_test

import (
	"fmt"
	"os"
	"testing"

	twitterscraper "github.com/masa-finance/masa-twitter-scraper"
)

var (
	username     = os.Getenv("TWITTER_USERNAME")
	password     = os.Getenv("TWITTER_PASSWORD")
	email        = os.Getenv("TWITTER_EMAIL")
	skipAuthTest = os.Getenv("SKIP_AUTH_TEST") != ""
	testScraper  = twitterscraper.New()
)

func init() {
	if username != "" && password != "" && !skipAuthTest {
		err := testScraper.Login(username, password, email)
		if err != nil {
			panic(fmt.Sprintf("Login() error = %v", err))
		}
	}
}

func TestAuth(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}
	scraper := twitterscraper.New()
	if err := scraper.Login(username, password, email); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !scraper.IsLoggedIn() {
		t.Fatalf("Expected IsLoggedIn() = true")
	}
	cookies := scraper.GetCookies()
	scraper2 := twitterscraper.New()
	scraper2.SetCookies(cookies)
	if !scraper2.IsLoggedIn() {
		t.Error("Expected restored IsLoggedIn() = true")
	}
	if err := scraper.Logout(); err != nil {
		t.Errorf("Logout() error = %v", err)
	}
	if scraper.IsLoggedIn() {
		t.Error("Expected IsLoggedIn() = false")
	}
}

func TestLoginOpenAccount(t *testing.T) {
	scraper := twitterscraper.New()
	if err := scraper.LoginOpenAccount(); err != nil {
		t.Fatalf("LoginOpenAccount() error = %v", err)
	}
}
