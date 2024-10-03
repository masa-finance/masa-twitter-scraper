package twitterscraper_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	twitterscraper "github.com/masa-finance/masa-twitter-scraper"
	"github.com/sirupsen/logrus"
)

var (
	username     string
	password     string
	email        string
	skipAuthTest bool
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	if err := godotenv.Load(); err != nil {
		logrus.WithError(err).Warn("Error loading .env file")
	}

	username = os.Getenv("TWITTER_USERNAME")
	password = os.Getenv("TWITTER_PASSWORD")
	email = os.Getenv("TWITTER_EMAIL")
	skipAuthTest = os.Getenv("SKIP_AUTH_TEST") != ""

	logrus.WithFields(logrus.Fields{
		"username":     username,
		"password":     password,
		"email":        email,
		"skipAuthTest": skipAuthTest,
	}).Info("Environment variables loaded")
}

func TestAuth(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	scraper := twitterscraper.New().SetHttpClient(getHTTPClientWithProxy())

	time.Sleep(100 * time.Millisecond)

	if err := scraper.Login(username, password, email); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !scraper.IsLoggedIn() {
		t.Fatalf("Expected IsLoggedIn() = true")
	}

	cookies := scraper.GetCookies()
	if err := saveCookiesToFile(cookies, "twitter_cookies.json"); err != nil {
		t.Fatalf("Failed to save cookies: %v", err)
	}

	t.Log("Successfully logged in and saved cookies")
}

func getHTTPClientWithProxy() *http.Client {
	proxyURL, err := url.Parse("http://sp7t5880k0:xv745uq1WpTI=veTrc@us.smartproxy.com:10001")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse proxy URL")
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
}

func saveCookiesToFile(cookies []*http.Cookie, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(cookies)
}
