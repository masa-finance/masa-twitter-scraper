package twitterscraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reHashtag    = regexp.MustCompile(`\B(\#\S+\b)`)
	reTwitterURL = regexp.MustCompile(`https:(\/\/t\.co\/([A-Za-z0-9]|[A-Za-z]){10})`)
	reUsername   = regexp.MustCompile(`\B(\@\S{1,15}\b)`)
	twURL        = urlParse("https://twitter.com")
)

func (s *Scraper) newRequest(method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("include_profile_interstitial_type", "1")
	q.Add("include_blocking", "1")
	q.Add("include_blocked_by", "1")
	q.Add("include_followed_by", "1")
	q.Add("include_want_retweets", "1")
	q.Add("include_mute_edge", "1")
	q.Add("include_can_dm", "1")
	q.Add("include_can_media_tag", "1")
	q.Add("include_ext_has_nft_avatar", "1")
	q.Add("include_ext_is_blue_verified", "1")
	q.Add("include_ext_verified_type", "1")
	q.Add("skip_status", "1")
	q.Add("cards_platform", "Web-12")
	q.Add("include_cards", "1")
	q.Add("include_ext_alt_text", "true")
	q.Add("include_ext_limited_action_results", "false")
	q.Add("include_quote_count", "true")
	q.Add("include_reply_count", "1")
	q.Add("tweet_mode", "extended")
	q.Add("include_ext_collab_control", "true")
	q.Add("include_ext_views", "true")
	q.Add("include_entities", "true")
	q.Add("include_user_entities", "true")
	q.Add("include_ext_media_color", "true")
	q.Add("include_ext_media_availability", "true")
	q.Add("include_ext_sensitive_media_warning", "true")
	q.Add("include_ext_trusted_friends_metadata", "true")
	q.Add("send_error_codes", "true")
	q.Add("simple_quoted_tweet", "true")
	q.Add("include_tweet_replies", strconv.FormatBool(s.includeReplies))
	q.Add("ext", "mediaStats,highlightedLabel,hasNftAvatar,voiceInfo,birdwatchPivot,enrichments,superFollowMetadata,unmentionInfo,editControl,collab_control,vibe")
	req.URL.RawQuery = q.Encode()

	return req, nil
}

func getUserTimeline(ctx context.Context, query string, maxProfilesNbr int, fetchFunc fetchProfileFunc) <-chan *ProfileResult {
	channel := make(chan *ProfileResult)
	go func(query string) {
		defer close(channel)
		var nextCursor string
		profilesNbr := 0
		for profilesNbr < maxProfilesNbr {
			select {
			case <-ctx.Done():
				channel <- &ProfileResult{Error: ctx.Err()}
				return
			default:
			}

			profiles, next, err := fetchFunc(query, maxProfilesNbr, nextCursor)
			if err != nil {
				channel <- &ProfileResult{Error: err}
				return
			}

			if len(profiles) == 0 {
				break
			}

			for _, profile := range profiles {
				select {
				case <-ctx.Done():
					channel <- &ProfileResult{Error: ctx.Err()}
					return
				default:
				}

				if profilesNbr < maxProfilesNbr {
					nextCursor = next
					channel <- &ProfileResult{Profile: *profile}
				} else {
					break
				}
				profilesNbr++
			}
		}
	}(query)
	return channel
}

func getTweetTimeline(ctx context.Context, query string, maxTweetsNbr int, fetchFunc fetchTweetFunc) <-chan *TweetResult {
	channel := make(chan *TweetResult)
	go func(query string) {
		defer close(channel)
		var nextCursor string
		tweetsNbr := 0
		for tweetsNbr < maxTweetsNbr {
			select {
			case <-ctx.Done():
				channel <- &TweetResult{Error: ctx.Err()}
				return
			default:
			}

			tweets, next, err := fetchFunc(query, maxTweetsNbr, nextCursor)
			if err != nil {
				channel <- &TweetResult{Error: err}
				return
			}

			if len(tweets) == 0 {
				break
			}

			for _, tweet := range tweets {
				select {
				case <-ctx.Done():
					channel <- &TweetResult{Error: ctx.Err()}
					return
				default:
				}

				if tweetsNbr < maxTweetsNbr {
					nextCursor = next
					channel <- &TweetResult{Tweet: *tweet}
				} else {
					break
				}
				tweetsNbr++
			}
		}
	}(query)
	return channel
}

func parseLegacyTweet(user *legacyUser, tweet *legacyTweet) *Tweet {
	tweetID := tweet.IDStr
	if tweetID == "" {
		return nil
	}
	username := user.ScreenName
	name := user.Name
	tw := &Tweet{
		ConversationID: tweet.ConversationIDStr,
		ID:             tweetID,
		Likes:          tweet.FavoriteCount,
		Name:           name,
		PermanentURL:   fmt.Sprintf("https://twitter.com/%s/status/%s", username, tweetID),
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
		tw.QuotedStatusID = tweet.QuotedStatusIDStr
	}
	if tweet.InReplyToStatusIDStr != "" {
		tw.IsReply = true
		tw.InReplyToStatusID = tweet.InReplyToStatusIDStr
	}
	if tweet.RetweetedStatusIDStr != "" || tweet.RetweetedStatusResult.Result != nil {
		tw.IsRetweet = true
		tw.RetweetedStatusID = tweet.RetweetedStatusIDStr
		if tweet.RetweetedStatusResult.Result != nil {
			tw.RetweetedStatus = parseLegacyTweet(&tweet.RetweetedStatusResult.Result.Core.UserResults.Result.Legacy, &tweet.RetweetedStatusResult.Result.Legacy)
			tw.RetweetedStatusID = tw.RetweetedStatus.ID
		}
	}

	if tweet.Views.Count != "" {
		views, viewsErr := strconv.Atoi(tweet.Views.Count)
		if viewsErr != nil {
			views = 0
		}
		tw.Views = views
	}

	for _, pinned := range user.PinnedTweetIdsStr {
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
		} else if media.Type == "animated_gif" {
			gif := GIF{
				ID:      media.IDStr,
				Preview: media.MediaURLHttps,
			}

			// Twitter's API doesn't provide bitrate for GIFs, (it's always set to zero).
			// Therefore we check for `>=` instead of `>` in the loop below.
			// Also, GIFs have just a single variant today. Just in case that changes in the future,
			// and there will be multiple variants, we'll pick the one with the highest bitrate,
			// if other one will have a non-zero bitrate.
			maxBitrate := 0
			for _, variant := range media.VideoInfo.Variants {
				if variant.Bitrate >= maxBitrate {
					gif.URL = variant.URL
					maxBitrate = variant.Bitrate
				}
			}

			tw.GIFs = append(tw.GIFs, gif)
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
	for _, gif := range tw.GIFs {
		url := gif.Preview
		if stringInSlice(url, foundedMedia) {
			continue
		}
		tw.HTML += fmt.Sprintf(`<br><img src="%s"/>`, url)
	}
	tw.HTML = strings.Replace(tw.HTML, "\n", "<br>", -1)
	return tw
}

func parseProfile(user legacyUser) Profile {
	profile := Profile{
		Avatar:         user.ProfileImageURLHTTPS,
		Banner:         user.ProfileBannerURL,
		Biography:      user.Description,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FavouritesCount,
		FriendsCount:   user.FriendsCount,
		IsPrivate:      user.Protected,
		IsVerified:     user.Verified,
		LikesCount:     user.FavouritesCount,
		ListedCount:    user.ListedCount,
		Location:       user.Location,
		Name:           user.Name,
		PinnedTweetIDs: user.PinnedTweetIdsStr,
		TweetsCount:    user.StatusesCount,
		URL:            "https://twitter.com/" + user.ScreenName,
		UserID:         user.IDStr,
		Username:       user.ScreenName,
	}

	tm, err := time.Parse(time.RubyDate, user.CreatedAt)
	if err == nil {
		tm = tm.UTC()
		profile.Joined = &tm
	}

	if len(user.Entities.URL.Urls) > 0 {
		profile.Website = user.Entities.URL.Urls[0].ExpandedURL
	}

	return profile
}

func mapToJSONString(data map[string]interface{}) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func urlParse(u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		return nil
	}
	return parsed
}
