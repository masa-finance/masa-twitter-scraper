package twitterscraper_test

import (
	"encoding/json"
	"fmt"
	"testing"
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

func parseLegacyInfo(jsonString string) ([]Legacy, error) {
	var response Response
	err := json.Unmarshal([]byte(jsonString), &response)
	if err != nil {
		return nil, err
	}

	var legacies []Legacy
	for _, instruction := range response.Data.User.Result.Timeline.Timeline.Instructions {
		for _, entry := range instruction.Entries {
			legacies = append(legacies, entry.Content.ItemContent.UserResults.Result.Legacy)
		}
	}

	return legacies, nil
}

func TestLegacyInfo(t *testing.T) {
	jsonString := `{"data": {"user": {"result": {"timeline": {"timeline": {"instructions": [{"entries": [{"content": {"itemContent": {"user_results": {"result": {"legacy": {"screen_name": "MayorFrancis","followers_count": 146754,"friends_count": 305,"listed_count": 1042,"created_at": "Mon Jun 25 00:19:02 +0000 2012","favourites_count": 28238,"statuses_count": 18092,"media_count": 2678,"profile_image_url_https": "https://pbs.twimg.com/profile_images/1696972096927047680/QxXgUhV__normal.jpg","description": "Proud to serve as @MiamiMayor. Former President of @USMayors.","location": "Miami, FL","url": "","protected": false,"verified": false}}}}}}]}}}}}}}}`

	legacies, err := parseLegacyInfo(jsonString)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	for _, legacy := range legacies {
		fmt.Printf("Legacy Info: %+v\n", legacy)
	}
}
