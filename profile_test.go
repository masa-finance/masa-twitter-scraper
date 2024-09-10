package twitterscraper_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/joho/godotenv"
	twitterscraper "github.com/masa-finance/masa-twitter-scraper"
)

func TestGetProfile(t *testing.T) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	username := os.Getenv("TWITTER_USERNAME")
	password := os.Getenv("TWITTER_PASSWORD")
	email := os.Getenv("TWITTER_EMAIL")
	skipAuthTest := os.Getenv("SKIP_AUTH_TEST") != ""

	if skipAuthTest {
		t.Skip("Skipping test due to SKIP_AUTH_TEST environment variable")
	}

	scraper := twitterscraper.New()
	if username != "" && password != "" {
		err := scraper.Login(username, password, email)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
	} else {
		t.Log("Running test without authentication. Some data might be limited.")
	}

	loc := time.FixedZone("UTC", 0)
	joined := time.Date(2010, 01, 18, 8, 49, 30, 0, loc)
	sample := twitterscraper.Profile{
		Avatar:         "https://pbs.twimg.com/profile_images/436075027193004032/XlDa2oaz_normal.jpeg",
		Banner:         "https://pbs.twimg.com/profile_banners/106037940/1541084318",
		Biography:      "nothing",
		IsPrivate:      false,
		IsVerified:     false,
		Joined:         &joined,
		Location:       "Ukraine",
		Name:           "Nomadic",
		PinnedTweetIDs: []string{},
		URL:            "https://twitter.com/nomadic_ua",
		UserID:         "106037940",
		Username:       "nomadic_ua",
		Website:        "https://nomadic.name",
	}

	profile, err := scraper.GetProfile("nomadic_ua")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	cmpOptions := cmp.Options{
		cmpopts.IgnoreFields(twitterscraper.Profile{}, "FollowersCount"),
		cmpopts.IgnoreFields(twitterscraper.Profile{}, "FollowingCount"),
		cmpopts.IgnoreFields(twitterscraper.Profile{}, "FriendsCount"),
		cmpopts.IgnoreFields(twitterscraper.Profile{}, "LikesCount"),
		cmpopts.IgnoreFields(twitterscraper.Profile{}, "ListedCount"),
		cmpopts.IgnoreFields(twitterscraper.Profile{}, "TweetsCount"),
	}
	if diff := cmp.Diff(sample, profile, cmpOptions...); diff != "" {
		t.Error("Resulting profile does not match the sample", diff)
	}

	if profile.FollowersCount == 0 {
		t.Error("Expected FollowersCount is greater than zero")
	}
	if profile.FollowingCount == 0 {
		t.Error("Expected FollowingCount is greater than zero")
	}
	if profile.LikesCount == 0 {
		t.Error("Expected LikesCount is greater than zero")
	}
	if profile.TweetsCount == 0 {
		t.Error("Expected TweetsCount is greater than zero")
	}
}
