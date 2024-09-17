package auth

import (
	"fmt"
	"strings"

	"github.com/masa-finance/masa-twitter-scraper/httpwrap"
	"github.com/masa-finance/masa-twitter-scraper/types"
)

const (
	LoginURL  = "https://api.twitter.com/1.1/onboarding/task.json"
	LogoutURL = "https://api.twitter.com/1.1/account/logout.json"
	OAuthURL  = "https://api.twitter.com/oauth2/token"

	SubtaskLoginJsInstrumentation = "LoginJsInstrumentationSubtask"
	SubtaskEnterUserIdentifier    = "LoginEnterUserIdentifierSSO"
	SubtaskEnterPassword          = "LoginEnterPassword"
	SubtaskAccountDuplication     = "AccountDuplicationCheck"
	SubtaskOpenAccount            = "OpenAccount"
)

// Login performs the authentication process using the provided LoginRequest.
// It manages the multi-step flow required for logging into a service that uses
// OAuth-like authentication mechanisms. The function handles various subtasks
// such as entering user credentials, handling two-factor authentication, and
// managing flow tokens.
//
// Parameters:
//   - request: A pointer to a LoginRequest struct containing all necessary
//     attributes for the login process, including bearer and guest tokens,
//     username, password, and optional confirmation data.
//
// Returns:
// - An error if any step in the login process fails.
//
// The function executes the following steps:
//  1. Initializes the login flow by setting up the initial data structure
//     required for the authentication process.
//  2. Handles JavaScript instrumentation, which may be required by the service
//     to track client-side interactions.
//  3. Submits the username as part of the login flow, ensuring the correct
//     user identifier is provided.
//  4. Submits the password to authenticate the user.
//  5. Checks for account duplication, which may occur if the account is already
//     logged in elsewhere.
//  6. Handles additional confirmation steps, such as two-factor authentication
//     or other security challenges, if required.
//
// Note:
//   - The function updates the FlowToken within the LoginRequest as the flow
//     progresses, ensuring the correct state is maintained across requests.
//   - If the login process requires additional confirmation (e.g., 2FA), the
//     Confirmation field in the LoginRequest must be populated.
//   - Ensure that sensitive information such as passwords and tokens are handled
//     securely and not exposed in logs or error messages.
func Login(request *LoginRequest) error {
	flowToken, err := startLoginFlow(request)
	if err != nil {
		return err
	}
	request.FlowToken = flowToken

	// Handle JavaScript instrumentation
	if err := handleSubtask(request, SubtaskLoginJsInstrumentation, map[string]interface{}{
		"js_instrumentation": map[string]interface{}{"response": "{}", "link": "next_link"},
	}); err != nil {
		return err
	}

	// Submit username
	if err := handleSubtask(request, SubtaskEnterUserIdentifier, map[string]interface{}{
		"settings_list": map[string]interface{}{
			"setting_responses": []map[string]interface{}{
				{
					"key":           "user_identifier",
					"response_data": map[string]interface{}{"text_data": map[string]interface{}{"result": request.Username}},
				},
			},
			"link": "next_link",
		},
	}); err != nil {
		return err
	}

	// Submit password
	if err := handleSubtask(request, SubtaskEnterPassword, map[string]interface{}{
		"enter_password": map[string]interface{}{"password": request.Password, "link": "next_link"},
	}); err != nil {
		return err
	}

	// Check for account duplication
	if err := handleSubtask(request, SubtaskAccountDuplication, map[string]interface{}{
		"check_logged_in_account": map[string]interface{}{"link": "AccountDuplicationCheck_false"},
	}); err != nil {
		return handleConfirmation(request, err)
	}
	return err
}

func startLoginFlow(request *LoginRequest) (string, error) {
	data := map[string]interface{}{
		"flow_name": "login",
		"input_flow_data": map[string]interface{}{
			"flow_context": map[string]interface{}{
				"debug_overrides": map[string]interface{}{},
				"start_location":  map[string]interface{}{"location": "splash_screen"},
			},
		},
	}
	return getFlowToken(data, request.BearerToken, request.GuestToken)
}

func handleSubtask(request *LoginRequest, subtaskID string, subtaskData map[string]interface{}) error {
	data := map[string]interface{}{
		"flow_token": request.FlowToken,
		"subtask_inputs": []map[string]interface{}{
			{
				"subtask_id": subtaskID,
			},
		},
	}
	for k, v := range subtaskData {
		data["subtask_inputs"].([]map[string]interface{})[0][k] = v
	}
	flowToken, err := getFlowToken(data, request.BearerToken, request.GuestToken)
	if err != nil {
		return err
	}
	request.FlowToken = flowToken
	return nil
}

func handleConfirmation(request *LoginRequest, err error) error {
	var confirmationSubtask string
	for _, subtask := range []string{"LoginAcid", "LoginTwoFactorAuthChallenge"} {
		if strings.Contains(err.Error(), subtask) {
			confirmationSubtask = subtask
			break
		}
	}
	if confirmationSubtask != "" {
		if request.Confirmation == "" {
			return fmt.Errorf("confirmation data required for %v", confirmationSubtask)
		}
		if err := handleSubtask(request, confirmationSubtask, map[string]interface{}{
			"enter_text": map[string]interface{}{"text": request.Confirmation, "link": "next_link"},
		}); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// LoginOpenAccount performs the process of opening an account using the provided OpenAccountRequest.
// It manages the multi-step flow required for opening an account, handling tasks such as
// obtaining access tokens and managing flow tokens.
//
// Parameters:
//   - request: A pointer to an OpenAccountRequest struct containing all necessary
//     attributes for the open account process, including bearer and guest tokens,
//     consumer keys, and secrets.
//
// Returns:
// - An error if any step in the open account process fails.
//
// The function executes the following steps:
//  1. Obtains an access token using the provided consumer key and secret.
//  2. Initializes the open account flow by setting up the initial data structure
//     required for the process.
//  3. Proceeds to the next task in the flow, which may involve opening a link or
//     performing additional actions.
//  4. Updates the OpenAccountRequest with the OAuth token and secret if the
//     process is successful.
//
// Note:
//   - The function updates the OpenAccountRequest with the OAuth token and secret,
//     as well as the login state, ensuring the correct state is maintained.
//   - Ensure that sensitive information such as tokens and secrets are handled
//     securely and not exposed in logs or error messages.
func LoginOpenAccount(request *OpenAccountRequest) error {
	accessToken, err := getAccessToken(request.ConsumerKey, request.ConsumerSecret)
	if err != nil {
		return err
	}
	request.BearerToken = accessToken

	// Flow start
	data := map[string]interface{}{
		"flow_name": "welcome",
		"input_flow_data": map[string]interface{}{
			"flow_context": map[string]interface{}{
				"debug_overrides": map[string]interface{}{},
				"start_location":  map[string]interface{}{"location": "splash_screen"},
			},
		},
	}
	flowToken, err := getFlowToken(data, request.BearerToken, request.GuestToken)
	if err != nil {
		return err
	}

	// Flow next link
	data = map[string]interface{}{
		"flow_token": flowToken,
		"subtask_inputs": []interface{}{
			map[string]interface{}{
				"subtask_id": "NextTaskOpenLink",
			},
		},
	}
	info, err := getFlow(data, request.BearerToken, request.GuestToken)
	if err != nil {
		return err
	}

	if info.Subtasks != nil && len(info.Subtasks) > 0 {
		if info.Subtasks[0].SubtaskID == "OpenAccount" {
			request.OAuthToken = info.Subtasks[0].OpenAccount.OAuthToken
			request.OAuthSecret = info.Subtasks[0].OpenAccount.OAuthTokenSecret
			if request.OAuthToken == "" || request.OAuthSecret == "" {
				return fmt.Errorf("auth error: %v", "Token or Secret is empty")
			}
			request.IsLogged = true
			request.IsOpenAccount = true
			return nil
		}
	}
	return fmt.Errorf("auth error: %v", "OpenAccount")
}

// getAccessToken retrieves an access token using the provided consumer key and secret.
// It sends a POST request to the OAuth URL with the necessary headers and data.
//
// Parameters:
// - consumerKey: The consumer key provided by the service for your application.
// - consumerSecret: The secret associated with the consumer key.
//
// Returns:
// - A string representing the access token.
// - An error if the request fails or the response cannot be parsed.
func getAccessToken(consumerKey, consumerSecret string) (string, error) {
	data := []byte("grant_type=client_credentials")
	headers := httpwrap.NewHeader()
	headers.AddContentType("application/x-www-form-urlencoded")
	headers.AddBasicAuth(consumerKey, consumerSecret)
	client := httpwrap.NewClient()

	var token types.Token
	result, _, err := client.Post(OAuthURL, data, headers, token)
	if err != nil {
		return "", err
	}
	token = result.(types.Token)
	return token.AccessToken, nil
}

// getFlow sends a POST request to the Login URL to retrieve flow information.
// It constructs the necessary headers and sends the request with the provided data.
//
// Parameters:
// - data: A map containing the flow data to be sent in the request body.
// - bearerToken: The bearer token for authentication.
// - guestToken: The guest token for authentication.
//
// Returns:
// - A pointer to a Flow struct containing the flow information.
// - An error if the request fails or the response cannot be parsed.
func getFlow(data map[string]interface{}, bearerToken, guestToken string) (*types.Flow, error) {
	headers := httpwrap.Header{
		"Authorization":             "Bearer " + bearerToken,
		"Content-Type":              "application/json",
		"User-Agent":                GetRandomUserAgent(),
		"X-Guest-Token":             guestToken,
		"X-Twitter-Auth-Type":       "OAuth2Client",
		"X-Twitter-Active-User":     "yes",
		"X-Twitter-Client-Language": "en",
	}
	var info types.Flow
	client := httpwrap.NewClient()
	result, _, err := client.Post(LoginURL, data, headers, info)
	if err != nil {
		return nil, err
	}
	info = result.(types.Flow)
	return &info, nil
}

// getFlowToken retrieves a flow token by sending a request to the flow endpoint.
// It processes the response to extract the flow token and handle any errors.
//
// Parameters:
// - data: A map containing the flow data to be sent in the request body.
// - bearerToken: The bearer token for authentication.
// - guestToken: The guest token for authentication.
//
// Returns:
// - A string representing the flow token.
// - An error if the request fails, the response contains errors, or the flow token cannot be extracted.
func getFlowToken(data map[string]interface{}, bearerToken, guestToken string) (string, error) {
	info, err := getFlow(data, bearerToken, guestToken)
	if err != nil {
		return "", err
	}
	if len(info.Errors) > 0 {
		return "", fmt.Errorf("auth error (%d): %v", info.Errors[0].Code, info.Errors[0].Message)
	}

	if info.Subtasks != nil && len(info.Subtasks) > 0 {
		if info.Subtasks[0].SubtaskID == "LoginEnterAlternateIdentifierSubtask" {
			err = fmt.Errorf("auth error: %v", "LoginEnterAlternateIdentifierSubtask")
		} else if info.Subtasks[0].SubtaskID == "LoginAcid" {
			err = fmt.Errorf("auth error: %v", "LoginAcid")
		} else if info.Subtasks[0].SubtaskID == "LoginTwoFactorAuthChallenge" {
			err = fmt.Errorf("auth error: %v", "LoginTwoFactorAuthChallenge")
		} else if info.Subtasks[0].SubtaskID == "DenyLoginSubtask" {
			err = fmt.Errorf("auth error: %v", "DenyLoginSubtask")
		}
	}
	return info.FlowToken, err
}
