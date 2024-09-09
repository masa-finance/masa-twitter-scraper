package twitterscraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// legacy timeline JSON object
type timelineV1 struct {
	GlobalObjects struct {
		Tweets map[string]legacyTweet `json:"tweets"`
		Users  map[string]legacyUser  `json:"users"`
	} `json:"globalObjects"`
	Timeline struct {
		Instructions []struct {
			AddEntries struct {
				Entries []struct {
					Content struct {
						Item struct {
							Content struct {
								Tweet struct {
									ID string `json:"id"`
								} `json:"tweet"`
								User struct {
									ID string `json:"id"`
								} `json:"user"`
							} `json:"content"`
						} `json:"item"`
						Operation struct {
							Cursor struct {
								Value      string `json:"value"`
								CursorType string `json:"cursorType"`
							} `json:"cursor"`
						} `json:"operation"`
						TimelineModule struct {
							Items []struct {
								Item struct {
									ClientEventInfo struct {
										Details struct {
											GuideDetails struct {
												TransparentGuideDetails struct {
													TrendMetadata struct {
														TrendName string `json:"trendName"`
													} `json:"trendMetadata"`
												} `json:"transparentGuideDetails"`
											} `json:"guideDetails"`
										} `json:"details"`
									} `json:"clientEventInfo"`
								} `json:"item"`
							} `json:"items"`
						} `json:"timelineModule"`
					} `json:"content,omitempty"`
				} `json:"entries"`
			} `json:"addEntries"`
			PinEntry struct {
				Entry struct {
					Content struct {
						Item struct {
							Content struct {
								Tweet struct {
									ID string `json:"id"`
								} `json:"tweet"`
							} `json:"content"`
						} `json:"item"`
					} `json:"content"`
				} `json:"entry"`
			} `json:"pinEntry,omitempty"`
			ReplaceEntry struct {
				Entry struct {
					Content struct {
						Operation struct {
							Cursor struct {
								Value      string `json:"value"`
								CursorType string `json:"cursorType"`
							} `json:"cursor"`
						} `json:"operation"`
					} `json:"content"`
				} `json:"entry"`
			} `json:"replaceEntry,omitempty"`
		} `json:"instructions"`
	} `json:"timeline"`
}

func (timeline *timelineV1) parseTweet(id string) *Tweet {
	if tweet, ok := timeline.GlobalObjects.Tweets[id]; ok {
		username := timeline.GlobalObjects.Users[tweet.UserIDStr].ScreenName
		name := timeline.GlobalObjects.Users[tweet.UserIDStr].Name
		tw := &Tweet{
			ID:             id,
			ConversationID: tweet.ConversationIDStr,
			Likes:          tweet.FavoriteCount,
			Name:           name,
			PermanentURL:   fmt.Sprintf("https://twitter.com/%s/status/%s", username, id),
			Replies:        tweet.ReplyCount,
			Retweets:       tweet.RetweetCount,
			Text:           tweet.FullText,
			UserID:         tweet.UserIDStr,
			Username:       username,
		}

		tm, err := time.Parse(time.RubyDate, tweet.CreatedAt)
		if err == nil {
			tw.TimeParsed = tm
			tw.Timestamp = tm.Unix()
		}

		if tweet.Place.ID != "" {
			tw.Place = &tweet.Place
		}

		if tweet.QuotedStatusIDStr != "" {
			tw.IsQuoted = true
			tw.QuotedStatus = timeline.parseTweet(tweet.QuotedStatusIDStr)
			tw.QuotedStatusID = tweet.QuotedStatusIDStr
		}
		if tweet.InReplyToStatusIDStr != "" {
			tw.IsReply = true
			tw.InReplyToStatus = timeline.parseTweet(tweet.InReplyToStatusIDStr)
			tw.InReplyToStatusID = tweet.InReplyToStatusIDStr
		}
		if tweet.RetweetedStatusIDStr != "" {
			tw.IsRetweet = true
			tw.RetweetedStatus = timeline.parseTweet(tweet.RetweetedStatusIDStr)
			tw.RetweetedStatusID = tweet.RetweetedStatusIDStr
		}

		if tweet.SelfThread.IDStr == id {
			tw.IsSelfThread = true
		}

		if tweet.Views.Count != "" {
			views, viewsErr := strconv.Atoi(tweet.Views.Count)
			if viewsErr != nil {
				views = 0
			}
			tw.Views = views
		}

		for _, pinned := range timeline.GlobalObjects.Users[tweet.UserIDStr].PinnedTweetIdsStr {
			if tweet.IDStr == pinned {
				tw.IsPin = true
				break
			}
		}

		for _, hash := range tweet.Entities.Hashtags {
			tw.Hashtags = append(tw.Hashtags, hash.Text)
		}

		for _, mention := range tweet.Entities.UserMentions {
			tw.Mentions = append(tw.Mentions, Mention{
				ID:       mention.IDStr,
				Username: mention.ScreenName,
				Name:     mention.Name,
			})
		}

		for _, media := range tweet.ExtendedEntities.Media {
			if media.Type == "photo" {
				photo := Photo{
					ID:  media.IDStr,
					URL: media.MediaURLHttps,
				}

				tw.Photos = append(tw.Photos, photo)
			} else if media.Type == "video" {
				video := Video{
					ID:      media.IDStr,
					Preview: media.MediaURLHttps,
				}

				maxBitrate := 0
				for _, variant := range media.VideoInfo.Variants {
					if variant.Bitrate > maxBitrate {
						video.URL = strings.TrimSuffix(variant.URL, "?tag=10")
						maxBitrate = variant.Bitrate
					}
				}

				tw.Videos = append(tw.Videos, video)
			}

			if !tw.SensitiveContent {
				sensitive := media.ExtSensitiveMediaWarning
				tw.SensitiveContent = sensitive.AdultContent || sensitive.GraphicViolence || sensitive.Other
			}
		}

		for _, url := range tweet.Entities.URLs {
			tw.URLs = append(tw.URLs, url.ExpandedURL)
		}

		tw.HTML = tweet.FullText
		tw.HTML = reHashtag.ReplaceAllStringFunc(tw.HTML, func(hashtag string) string {
			return fmt.Sprintf(`<a href="https://twitter.com/hashtag/%s">%s</a>`,
				strings.TrimPrefix(hashtag, "#"),
				hashtag,
			)
		})
		tw.HTML = reUsername.ReplaceAllStringFunc(tw.HTML, func(username string) string {
			return fmt.Sprintf(`<a href="https://twitter.com/%s">%s</a>`,
				strings.TrimPrefix(username, "@"),
				username,
			)
		})
		var foundedMedia []string
		tw.HTML = reTwitterURL.ReplaceAllStringFunc(tw.HTML, func(tco string) string {
			for _, entity := range tweet.Entities.URLs {
				if tco == entity.URL {
					return fmt.Sprintf(`<a href="%s">%s</a>`, entity.ExpandedURL, tco)
				}
			}
			for _, entity := range tweet.ExtendedEntities.Media {
				if tco == entity.URL {
					foundedMedia = append(foundedMedia, entity.MediaURLHttps)
					return fmt.Sprintf(`<br><a href="%s"><img src="%s"/></a>`, tco, entity.MediaURLHttps)
				}
			}
			return tco
		})
		for _, photo := range tw.Photos {
			url := photo.URL
			if stringInSlice(url, foundedMedia) {
				continue
			}
			tw.HTML += fmt.Sprintf(`<br><img src="%s"/>`, url)
		}
		for _, video := range tw.Videos {
			url := video.Preview
			if stringInSlice(url, foundedMedia) {
				continue
			}
			tw.HTML += fmt.Sprintf(`<br><img src="%s"/>`, url)
		}
		tw.HTML = strings.Replace(tw.HTML, "\n", "<br>", -1)
		return tw
	}
	return nil
}

func (timeline *timelineV1) parseTweets() ([]*Tweet, string) {
	var cursor string
	var pinnedTweet *Tweet
	var orderedTweets []*Tweet
	for _, instruction := range timeline.Timeline.Instructions {
		if instruction.PinEntry.Entry.Content.Item.Content.Tweet.ID != "" {
			if tweet := timeline.parseTweet(instruction.PinEntry.Entry.Content.Item.Content.Tweet.ID); tweet != nil {
				pinnedTweet = tweet
			}
		}
		for _, entry := range instruction.AddEntries.Entries {
			if tweet := timeline.parseTweet(entry.Content.Item.Content.Tweet.ID); tweet != nil {
				orderedTweets = append(orderedTweets, tweet)
			}
			if entry.Content.Operation.Cursor.CursorType == "Bottom" {
				cursor = entry.Content.Operation.Cursor.Value
			}
		}
		if instruction.ReplaceEntry.Entry.Content.Operation.Cursor.CursorType == "Bottom" {
			cursor = instruction.ReplaceEntry.Entry.Content.Operation.Cursor.Value
		}
	}
	if pinnedTweet != nil && len(orderedTweets) > 0 {
		orderedTweets = append([]*Tweet{pinnedTweet}, orderedTweets...)
	}
	return orderedTweets, cursor
}

func (timeline *timelineV1) parseUsers() ([]*Profile, string) {
	users := make(map[string]Profile)

	for id, user := range timeline.GlobalObjects.Users {
		users[id] = parseProfile(user)
	}

	var cursor string
	var orderedProfiles []*Profile
	for _, instruction := range timeline.Timeline.Instructions {
		for _, entry := range instruction.AddEntries.Entries {
			if profile, ok := users[entry.Content.Item.Content.User.ID]; ok {
				orderedProfiles = append(orderedProfiles, &profile)
			}
			if entry.Content.Operation.Cursor.CursorType == "Bottom" {
				cursor = entry.Content.Operation.Cursor.Value
			}
		}
		if instruction.ReplaceEntry.Entry.Content.Operation.Cursor.CursorType == "Bottom" {
			cursor = instruction.ReplaceEntry.Entry.Content.Operation.Cursor.Value
		}
	}
	return orderedProfiles, cursor
}
