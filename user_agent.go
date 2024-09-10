package twitterscraper

import (
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
)

var (
	userAgents       []string
	currentUserAgent string
	userAgentMutex   sync.Mutex
)

func init() {
	if err := godotenv.Load(); err != nil {
		logrus.WithError(err).Warn("Error loading .env file")
	}
	loadUserAgents()
}

func loadUserAgents() {
	userAgentString := os.Getenv("USER_AGENTS")
	logrus.WithField("raw_user_agents", userAgentString).Debug("Raw USER_AGENTS value")

	if userAgentString == "" {
		logrus.Warn("USER_AGENTS environment variable is not set. Using default user agent.")
		userAgents = []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"}
	} else {
		// Use a more robust splitting method
		userAgents = splitUserAgents(userAgentString)
		logrus.WithFields(logrus.Fields{
			"count":  len(userAgents),
			"agents": userAgents,
		}).Info("Loaded user agents from environment")
	}
}

func splitUserAgents(s string) []string {
	var result []string
	var builder strings.Builder
	inParentheses := false

	for _, r := range s {
		switch r {
		case '(':
			inParentheses = true
			builder.WriteRune(r)
		case ')':
			inParentheses = false
			builder.WriteRune(r)
		case ',':
			if inParentheses {
				builder.WriteRune(r)
			} else {
				if builder.Len() > 0 {
					result = append(result, strings.TrimSpace(builder.String()))
					builder.Reset()
				}
			}
		default:
			builder.WriteRune(r)
		}
	}

	if builder.Len() > 0 {
		result = append(result, strings.TrimSpace(builder.String()))
	}

	return result
}

func GetUserAgent() string {
	userAgentMutex.Lock()
	defer userAgentMutex.Unlock()

	if currentUserAgent == "" {
		currentUserAgent = userAgents[0]
		logrus.WithField("user_agent", currentUserAgent).Debug("Using first user agent from list")
	}
	return currentUserAgent
}

func GetRandomUserAgent() string {
	userAgentMutex.Lock()
	defer userAgentMutex.Unlock()

	newAgent := userAgents[rand.Intn(len(userAgents))]
	if newAgent != currentUserAgent {
		currentUserAgent = newAgent
		logrus.WithField("user_agent", currentUserAgent).Debug("Using new random user agent")
	}
	return currentUserAgent
}
