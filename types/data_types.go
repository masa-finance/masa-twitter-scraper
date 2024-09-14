package types

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Token struct {
	AccessToken string `json:"access_token"`
}

type Flow struct {
	Errors    []Error `json:"errors"`
	FlowToken string  `json:"flow_token"`
	Status    string  `json:"status"`
	Subtasks  []struct {
		SubtaskID   string `json:"subtask_id"`
		OpenAccount struct {
			OAuthToken       string `json:"oauth_token"`
			OAuthTokenSecret string `json:"oauth_token_secret"`
		} `json:"open_account"`
	} `json:"subtasks"`
}

type VerifyCredentials struct {
	Errors []Error `json:"errors"`
}
