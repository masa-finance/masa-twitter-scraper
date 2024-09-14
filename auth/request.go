package auth

// LoginRequest contains all the necessary attributes for performing a login.
type LoginRequest struct {
	BearerToken  string
	GuestToken   string
	Username     string
	Password     string
	Confirmation string
	FlowToken    string
}

// NewLoginRequest creates a new LoginRequest with the provided credentials and tokens.
func NewLoginRequest(bearerToken, guestToken, username, password, confirmation string) *LoginRequest {
	return &LoginRequest{
		BearerToken:  bearerToken,
		GuestToken:   guestToken,
		Username:     username,
		Password:     password,
		Confirmation: confirmation,
	}
}

// OpenAccountRequest contains all the necessary attributes for opening an account.
type OpenAccountRequest struct {
	BearerToken    string
	GuestToken     string
	ConsumerKey    string
	ConsumerSecret string
	OAuthToken     string
	OAuthSecret    string
	IsLogged       bool
	IsOpenAccount  bool
}

// NewOpenAccountRequest creates a new OpenAccountRequest with the provided tokens and keys.
func NewOpenAccountRequest(bearerToken, guestToken, consumerKey, consumerSecret string) *OpenAccountRequest {
	return &OpenAccountRequest{
		BearerToken:    bearerToken,
		GuestToken:     guestToken,
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
	}
}
