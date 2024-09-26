package twitterscraper_test

import (
	"encoding/json"
	"net/http"
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
}

func TestAuth(t *testing.T) {
	if skipAuthTest {
		t.Skip("Skipping test due to environment variable")
	}

	scraper := twitterscraper.New()

	// Add a short delay before login attempt
	time.Sleep(2 * time.Second)

	err := scraper.Login(username, password, email)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	// Add a short delay after login attempt
	time.Sleep(2 * time.Second)

	if !scraper.IsLoggedIn() {
		t.Fatalf("Expected IsLoggedIn() = true")
	}

	// Save cookies
	cookies := scraper.GetCookies()
	err = saveCookiesToFile(cookies, "twitter_cookies.json")
	if err != nil {
		t.Fatalf("Failed to save cookies: %v", err)
	}

	// Log success
	t.Log("Successfully logged in and saved cookies")
}

func saveCookiesToFile(cookies []*http.Cookie, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(cookies)
}
