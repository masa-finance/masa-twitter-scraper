package twitterscraper

import (
	"net/url"
)

type FollowingResponse struct {
	Data Data `json:"data"`
}

type Data struct {
	User User `json:"user"`
}

type User struct {
	Result UserResult `json:"result"`
}

type UserResult struct {
	Timeline UserTimeline `json:"timeline"`
}

type UserTimeline struct {
	Timeline InnerTimeline `json:"timeline"`
}

type InnerTimeline struct {
	Instructions []Instruction `json:"instructions"`
}

type Instruction struct {
	Type       string  `json:"type"`
	Direction  string  `json:"direction,omitempty"`
	Entries    []Entry `json:"entries,omitempty"`
	Value      string  `json:"value,omitempty"`
	CursorType string  `json:"cursorType,omitempty"`
}

type Entry struct {
	EntryId   string  `json:"entryId"`
	SortIndex string  `json:"sortIndex"`
	Content   Content `json:"content"`
}

type Content struct {
	EntryType       string           `json:"entryType"`
	ItemType        string           `json:"itemType,omitempty"`
	UserResults     *UserResults     `json:"user_results,omitempty"`
	UserDisplayType string           `json:"userDisplayType,omitempty"`
	ClientEventInfo *ClientEventInfo `json:"clientEventInfo,omitempty"`
}

type UserResults struct {
	Result UserProfile `json:"result"`
}

type UserProfile struct {
	ID             string         `json:"id"`
	RestID         string         `json:"rest_id"`
	Legacy         Legacy         `json:"legacy"`
	Professional   Professional   `json:"professional,omitempty"`
	TipjarSettings TipjarSettings `json:"tipjar_settings,omitempty"`
}

type Legacy struct {
	CanDM                bool   `json:"can_dm"`
	Description          string `json:"description"`
	Name                 string `json:"name"`
	ScreenName           string `json:"screen_name"`
	FollowersCount       int    `json:"followers_count"`
	FriendsCount         int    `json:"friends_count"`
	StatusesCount        int    `json:"statuses_count"`
	ProfileImageURLHttps string `json:"profile_image_url_https"`
	Verified             bool   `json:"verified"`
	// Add more fields as necessary
}

type Professional struct {
	RestID           string     `json:"rest_id"`
	ProfessionalType string     `json:"professional_type"`
	Category         []Category `json:"category,omitempty"`
}

type Category struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IconName string `json:"icon_name"`
}

type TipjarSettings struct {
	IsEnabled     bool   `json:"is_enabled,omitempty"`
	CashAppHandle string `json:"cash_app_handle,omitempty"`
	PatreonHandle string `json:"patreon_handle,omitempty"`
}

type ClientEventInfo struct {
	Component string `json:"component"`
	Element   string `json:"element"`
}

// FetchFollowers gets the list of followers for a given user, via the Twitter frontend GraphQL API.
func (s *Scraper) FetchFollowers(userID string, maxUsersNbr int, cursor string) ([]*Legacy, string, error) {
	if maxUsersNbr > 200 {
		maxUsersNbr = 200
	}

	req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/o1YfmoGa-hb8Z6yQhoIBhg/Followers")
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"userId":                 userID,
		"count":                  maxUsersNbr,
		"includePromotedContent": false,
	}
	features := map[string]interface{}{"rweb_tipjar_consumption_enabled": true, "responsive_web_graphql_exclude_directive_enabled": true, "verified_phone_label_enabled": false, "creator_subscriptions_tweet_preview_api_enabled": true, "responsive_web_graphql_timeline_navigation_enabled": true, "responsive_web_graphql_skip_user_profile_image_extensions_enabled": false, "communities_web_enable_tweet_community_results_fetch": true, "c9s_tweet_anatomy_moderator_badge_enabled": true, "articles_preview_enabled": true, "tweetypie_unmention_optimization_enabled": true, "responsive_web_edit_tweet_api_enabled": true, "graphql_is_translatable_rweb_tweet_is_translatable_enabled": true, "view_counts_everywhere_api_enabled": true, "longform_notetweets_consumption_enabled": true, "responsive_web_twitter_article_tweet_consumption_enabled": true, "tweet_awards_web_tipping_enabled": false, "creator_subscriptions_quote_tweet_preview_enabled": false, "freedom_of_speech_not_reach_fetch_enabled": true, "standardized_nudges_misinfo": true, "tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true, "tweet_with_visibility_results_prefer_gql_media_interstitial_enabled": true, "rweb_video_timestamps_enabled": true, "longform_notetweets_rich_text_read_enabled": true, "longform_notetweets_inline_media_enabled": true, "responsive_web_enhance_cards_enabled": false}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(features))
	req.URL.RawQuery = query.Encode()

	var response FollowingResponse // You might need to adjust the response struct if the followers response differs
	err = s.RequestAPI(req, &response)
	if err != nil {
		return nil, "", err
	}

	legacies, nextCursor, err := response.parseFollowing() // Capture the error value as well
	if err != nil {
		return nil, "", err // Handle the error appropriately
	}
	return legacies, nextCursor, nil
}

func (fr *FollowingResponse) parseFollowing() ([]*Legacy, string, error) {
	var legacies []*Legacy
	var nextCursor string

	for _, instruction := range fr.Data.User.Result.Timeline.Timeline.Instructions {
		if instruction.Type == "TimelineAddEntries" {
			for _, entry := range instruction.Entries {
				if entry.Content.EntryType == "User" {
					legacies = append(legacies, &entry.Content.UserResults.Result.Legacy)
				}
			}
		} else if instruction.Type == "TimelinePinEntry" {
			nextCursor = instruction.Value
		}
	}

	return legacies, nextCursor, nil
}
