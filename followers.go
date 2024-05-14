package twitterscraper

import (
	"log"
	"net/url"
)

type Response struct {
	Data struct {
		User struct {
			Result struct {
				Timeline struct {
					Timeline struct {
						Instructions []struct {
							Entries []struct {
								Content struct {
									ItemContent struct {
										UserResults struct {
											Result struct {
												Legacy Legacy `json:"legacy"`
											} `json:"result"`
										} `json:"user_results"`
									} `json:"itemContent"`
								} `json:"content"`
							} `json:"entries"`
						} `json:"instructions"`
					} `json:"timeline"`
				} `json:"timeline"`
			} `json:"result"`
		} `json:"user"`
	} `json:"data"`
}

type Legacy struct {
	ScreenName      string `json:"screen_name"`
	FollowersCount  int    `json:"followers_count"`
	FriendsCount    int    `json:"friends_count"`
	ListedCount     int    `json:"listed_count"`
	CreatedAt       string `json:"created_at"`
	FavouritesCount int    `json:"favourites_count"`
	StatusesCount   int    `json:"statuses_count"`
	MediaCount      int    `json:"media_count"`
	ProfileImageUrl string `json:"profile_image_url_https"`
	Description     string `json:"description"`
	Location        string `json:"location"`
	Url             string `json:"url"`
	Protected       bool   `json:"protected"`
	Verified        bool   `json:"verified"`
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

	var response Response
	err = s.RequestAPI(req, &response)
	if err != nil {
		// Handle the error, for example, log it or return it to the caller
		log.Printf("Error making API request: %v", err)
		return nil, "", err
	}

	legacies, nextCursor, err := response.parseFollowing()
	if err != nil {
		// Handle the parsing error
		log.Printf("Error parsing following response: %v", err)
		return nil, "", err
	}

	// If err is nil here, it means both the API request and parsing were successful
	return legacies, nextCursor, nil
}

func (fr Response) parseFollowing() ([]*Legacy, string, error) {
	var legacies []*Legacy

	log.Println("Starting to parse following...") // Log the start of the parsing process

	for _, instruction := range fr.Data.User.Result.Timeline.Timeline.Instructions {
		log.Printf("Processing instruction with %d entries\n", len(instruction.Entries)) // Log the number of entries in the current instruction
		for _, entry := range instruction.Entries {
			// Append the address of Legacy struct to the slice
			legacies = append(legacies, &entry.Content.ItemContent.UserResults.Result.Legacy)
			log.Printf("Added legacy for user: %s\n", entry.Content.ItemContent.UserResults.Result.Legacy.ScreenName) // Log the screen name of the user being added
		}
	}

	// Assuming the next cursor is part of your response, you need to extract it here.
	// This is a placeholder for where you would extract the cursor from your response.
	// Adjust this according to your actual JSON structure.
	nextCursor := "" // Placeholder: Extract the actual cursor from the response
	log.Println("Parsing complete. No next cursor extracted.") // Log completion and note about the cursor

	return legacies, nextCurs
