package twitterscraper

import (
	"context"
	"errors"
	"net/url"
	"strconv"
)

const searchURL = "https://twitter.com/i/api/graphql/MJpyQGqgklrVl_0X9gNy3A/SearchTimeline"

type searchTimeline struct {
	Data struct {
		SearchByRawQuery struct {
			SearchTimeline struct {
				Timeline struct {
					Instructions []struct {
						Type    string  `json:"type"`
						Entries []entry `json:"entries"`
						Entry   entry   `json:"entry,omitempty"`
					} `json:"instructions"`
				} `json:"timeline"`
			} `json:"search_timeline"`
		} `json:"search_by_raw_query"`
	} `json:"data"`
}

func (timeline *searchTimeline) parseTweets() ([]*Tweet, string) {
	tweets := make([]*Tweet, 0)
	cursor := ""
	for _, instruction := range timeline.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions {
		if instruction.Type == "TimelineAddEntries" || instruction.Type == "TimelineReplaceEntry" {
			if instruction.Entry.Content.CursorType == "Bottom" {
				cursor = instruction.Entry.Content.Value
				continue
			}
			for _, entry := range instruction.Entries {
				if entry.Content.ItemContent.TweetDisplayType == "Tweet" {
					if tweet := parseLegacyTweet(&entry.Content.ItemContent.TweetResults.Result.Core.UserResults.Result.Legacy, &entry.Content.ItemContent.TweetResults.Result.Legacy); tweet != nil {
						if tweet.Views == 0 && entry.Content.ItemContent.TweetResults.Result.Views.Count != "" {
							tweet.Views, _ = strconv.Atoi(entry.Content.ItemContent.TweetResults.Result.Views.Count)
						}
						tweets = append(tweets, tweet)
					}
				} else if entry.Content.CursorType == "Bottom" {
					cursor = entry.Content.Value
				}
			}
		}
	}
	return tweets, cursor
}

func (timeline *searchTimeline) parseUsers() ([]*Profile, string) {
	profiles := make([]*Profile, 0)
	cursor := ""
	for _, instruction := range timeline.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions {
		if instruction.Type == "TimelineAddEntries" || instruction.Type == "TimelineReplaceEntry" {
			if instruction.Entry.Content.CursorType == "Bottom" {
				cursor = instruction.Entry.Content.Value
				continue
			}
			for _, entry := range instruction.Entries {
				if entry.Content.ItemContent.UserDisplayType == "User" {
					if profile := parseProfile(entry.Content.ItemContent.UserResults.Result.Legacy); profile.Name != "" {
						if profile.UserID == "" {
							profile.UserID = entry.Content.ItemContent.UserResults.Result.RestID
						}
						profiles = append(profiles, &profile)
					}
				} else if entry.Content.CursorType == "Bottom" {
					cursor = entry.Content.Value
				}
			}
		}
	}
	return profiles, cursor
}

// SearchTweets returns channel with tweets for a given search query
func (s *Scraper) SearchTweets(ctx context.Context, query string, maxTweetsNbr int) <-chan *TweetResult {
	return getTweetTimeline(ctx, query, maxTweetsNbr, s.FetchSearchTweets)
}

// SearchProfiles returns channel with profiles for a given search query
func (s *Scraper) SearchProfiles(ctx context.Context, query string, maxProfilesNbr int) <-chan *ProfileResult {
	return getUserTimeline(ctx, query, maxProfilesNbr, s.FetchSearchProfiles)
}

// getSearchTimeline gets results for a given search query, via the Twitter frontend API
func (s *Scraper) getSearchTimeline(query string, maxNbr int, cursor string) (*searchTimeline, error) {
	if !s.isLogged {
		return nil, errors.New("scraper is not logged in for search")
	}

	if maxNbr > 50 {
		maxNbr = 50
	}

	req, err := s.newRequest("GET", searchURL)
	if err != nil {
		return nil, err
	}

	// Add all required headers
	req.Header.Set("authority", "twitter.com")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("referer", "https://twitter.com/search?q="+url.QueryEscape(query)+"&src=typed_query&f=top")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "macOS")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-client-language", "en")

	// Update features map with all required fields
	features := map[string]interface{}{
		"rweb_tipjar_consumption_enabled":                                         true,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"communities_web_enable_tweet_community_results_fetch":                    true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"articles_preview_enabled":                                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"creator_subscriptions_quote_tweet_preview_enabled":                       false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"rweb_video_timestamps_enabled":                                           true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	// Add fieldToggles map
	fieldToggles := map[string]interface{}{
		"withArticleRichContentState": false,
	}

	variables := map[string]interface{}{
		"rawQuery":    query,
		"count":       maxNbr,
		"querySource": "typed_query",
		"product":     "Top",
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	// Update query parameters to include fieldToggles
	q := url.Values{}
	q.Set("variables", mapToJSONString(variables))
	q.Set("features", mapToJSONString(features))
	q.Set("fieldToggles", mapToJSONString(fieldToggles))
	req.URL.RawQuery = q.Encode()

	var timeline searchTimeline
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, err
	}
	return &timeline, nil
}

// FetchSearchTweets gets tweets for a given search query, via the Twitter frontend API
func (s *Scraper) FetchSearchTweets(query string, maxTweetsNbr int, cursor string) ([]*Tweet, string, error) {
	timeline, err := s.getSearchTimeline(query, maxTweetsNbr, cursor)
	if err != nil {
		return nil, "", err
	}
	tweets, nextCursor := timeline.parseTweets()
	return tweets, nextCursor, nil
}

// FetchSearchProfiles gets users for a given search query, via the Twitter frontend API
func (s *Scraper) FetchSearchProfiles(query string, maxProfilesNbr int, cursor string) ([]*Profile, string, error) {
	timeline, err := s.getSearchTimeline(query, maxProfilesNbr, cursor)
	if err != nil {
		return nil, "", err
	}
	users, nextCursor := timeline.parseUsers()
	return users, nextCursor, nil
}
