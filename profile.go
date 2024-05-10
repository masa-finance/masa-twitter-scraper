package twitterscraper

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Global cache for user IDs
var cacheIDs sync.Map

// Profile of twitter user.
type Profile struct {
	Avatar         string
	Banner         string
	Biography      string
	Birthday       string
	FollowersCount int
	FollowingCount int
	FriendsCount   int
	IsPrivate      bool
	IsVerified     bool
	Joined         *time.Time
	LikesCount     int
	ListedCount    int
	Location       string
	Name           string
	PinnedTweetIDs []string
	TweetsCount    int
	URL            string
	UserID         string
	Username       string
	Website        string
}

type user struct {
	Data struct {
		User struct {
			RestID string     `json:"rest_id"`
			Legacy legacyUser `json:"legacy"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// GetProfile return parsed user profile.
func (s *Scraper) GetProfile(username string) (Profile, error) {
	var jsn user
	req, err := http.NewRequest("GET", "https://api.twitter.com/graphql/4S2ihIKfF3xhp-ENxvUAfQ/UserByScreenName", nil)
	if err != nil {
		return Profile{}, err
	}

	variables := map[string]interface{}{
		"screen_name":          username,
		"withHighlightedLabel": true,
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	req.URL.RawQuery = query.Encode()

	err = s.RequestAPI(req, &jsn)
	if err != nil {
		return Profile{}, err
	}

	if len(jsn.Errors) > 0 {
		return Profile{}, fmt.Errorf("%s", jsn.Errors[0].Message)
	}

	if jsn.Data.User.RestID == "" {
		return Profile{}, fmt.Errorf("rest_id not found")
	}
	jsn.Data.User.Legacy.IDStr = jsn.Data.User.RestID

	if jsn.Data.User.Legacy.ScreenName == "" {
		return Profile{}, fmt.Errorf("either @%s does not exist or is private", username)
	}

	return parseProfile(jsn.Data.User.Legacy), nil
}

// GetUserIDByScreenName from API
func (s *Scraper) GetUserIDByScreenName(screenName string) (string, error) {
	id, ok := cacheIDs.Load(screenName)
	if ok {
		return id.(string), nil
	}

	profile, err := s.GetProfile(screenName)
	if err != nil {
		return "", err
	}

	cacheIDs.Store(screenName, profile.UserID)

	return profile.UserID, nil
}
