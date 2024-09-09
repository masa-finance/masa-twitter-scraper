package twitterscraper

import "golang.org/x/exp/rand"

var UserAgents = []string{
	// Chrome on Mac OS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
}

func GetUserAgent() string {
	return UserAgents[0]
}

func GetRandomUserAgent() string {
	return UserAgents[rand.Intn(len(UserAgents))]
}
