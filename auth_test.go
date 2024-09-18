package twitterscraper_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	twitterscraper "github.com/masa-finance/masa-twitter-scraper"
	"github.com/sirupsen/logrus"
)

var (
	username     string
	password     string
	email        string
	skipAuthTest bool
	testScraper  = twitterscraper.New()
)

func init() {
	// Set log level to Debug
	logrus.SetLevel(logrus.DebugLevel)

	// Optionally, set a custom formatter
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

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

	if username != "" && password != "" && !skipAuthTest {
		err := testScraper.Login(username, password, email)
		if err != nil {
			logrus.WithError(err).Panic("Login failed")
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
