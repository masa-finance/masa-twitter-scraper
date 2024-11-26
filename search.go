package twitterscraper

import (
	"context"
	"net/http"
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
func (s *Scraper) getSearchTimeline(query string, maxTweetsNbr int, cursor string) (*searchTimeline, error) {
	req, err := http.NewRequest("GET", "https://x.com/i/api/graphql/MJpyQGqgklrVl_0X9gNy3A/SearchTimeline", nil)
	if err != nil {
		return nil, err
	}

	// Set required headers
	req.Header.Set("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-client-language", "en")
	req.Header.Set("accept", "*/*")
	req.Header.Set("referer", "https://x.com/search?q="+url.QueryEscape(query)+"&src=typed_query")

	// Set variables
	variables := map[string]interface{}{
		"rawQuery":    query,
		"count":       maxTweetsNbr,
		"querySource": "typed_query",
		"product":     "Top",
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	// Set features
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

	// Build query parameters
	q := url.Values{}
	q.Set("variables", mapToJSONString(variables))
	q.Set("features", mapToJSONString(features))
	req.URL.RawQuery = q.Encode()

	// Set CSRF token from cookies
	if cookies := s.client.Jar.Cookies(req.URL); len(cookies) > 0 {
		for _, cookie := range cookies {
			if cookie.Name == "ct0" {
				req.Header.Set("x-csrf-token", cookie.Value)
				break
			}
		}
	}

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
