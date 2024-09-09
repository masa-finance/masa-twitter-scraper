package twitterscraper

const (
	loginURL             = "https://api.twitter.com/1.1/onboarding/task.json"
	logoutURL            = "https://api.twitter.com/1.1/account/logout.json"
	oAuthURL             = "https://api.twitter.com/oauth2/token"
	verifyCredentialsURL = "https://api.twitter.com/1.1/account/verify_credentials.json"
	bearerToken2         = "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
	appConsumerKey       = "3nVuSoBZnx6U4vzUxf5w"
	appConsumerSecret    = "Bcs59EFbbsdF6Sl9Ng71smgStWEGwXXKSjYvPVt7qys"
)

type (
	flow struct {
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		FlowToken string `json:"flow_token"`
		Status    string `json:"status"`
		Subtasks  []struct {
			SubtaskID   string `json:"subtask_id"`
			OpenAccount struct {
				OAuthToken       string `json:"oauth_token"`
				OAuthTokenSecret string `json:"oauth_token_secret"`
			} `json:"open_account"`
		} `json:"subtasks"`
	}

	verifyCredentials struct {
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}
)
