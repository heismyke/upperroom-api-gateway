package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lambda-func/database"
	"lambda-func/types"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
  _ "github.com/jpfuentes2/go-env/autoload"
	"github.com/aws/aws-lambda-go/events"
)

type ApiHandler struct {
	tokenStore    database.AuthStore
	httpClient    *http.Client
  clientID string
  clientSecret string
  redirect_uri string
	monitorCtx    context.Context
	monitorCancel context.CancelFunc
}

var(
  clientID = os.Getenv("CLIENT_ID")
  clientSecret = os.Getenv("CLIENT_SECRET")
  redirect_uri = os.Getenv("REDIRECT_URI")
)

func NewApiHandler(tokenStore database.AuthStore) *ApiHandler {
// Log configuration on startup
    log.Printf("Initializing API Handler with client_id length: %d", len(clientID))
    if clientID == "" {
        log.Printf("WARNING: CLIENT_ID environment variable is not set")
    }
    if clientSecret == "" {
        log.Printf("WARNING: CLIENT_SECRET environment variable is not set")
    }
    if os.Getenv("REDIRECT_URI") == "" {
        log.Printf("WARNING: REDIRECT_URI environment variable is not set")
    }
	ctx, cancel := context.WithCancel(context.Background())
	return &ApiHandler{
		tokenStore:    tokenStore,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
    clientID : clientID,
    clientSecret: clientSecret,
    redirect_uri: redirect_uri,

		monitorCtx:    ctx,
		monitorCancel: cancel,
	}
}

func (api *ApiHandler) StoreTokenHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  log.Printf("Received request: %+v", request) // Logs the full request

    if request.Path != "/callback" {
        return events.APIGatewayProxyResponse{
            StatusCode: 404,
            Body:       `{"error": "Not Found"}`,
            Headers:    map[string]string{"Content-Type": "application/json"},
        }, nil
    }

    log.Printf("QueryStringParameters: %+v", request.QueryStringParameters)
	authCode, exists := request.QueryStringParameters["code"]
	if !exists || len(authCode) == 0 {
		log.Println("Error: Missing or empty authCode")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Missing authorization code in callback"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	log.Printf("Received authCode: '%s' (length: %d)", authCode, len(authCode))

	// Ensure you are not slicing an empty string
	if len(authCode) < 4 {
		log.Println("Error: authCode is too short to process")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Invalid authorization code"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	accessToken, err := api.exchangeCodeForToken(authCode)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error": "Failed to exchange token: %v"}`, err),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	if err := api.tokenStore.StoreToken(accessToken); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
      Body:       fmt.Sprintf(`{"error": "Failed to store token: %v"}`, err),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}


	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"message": "Success! Monitoring for 'Upper Room' concerts"}`,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func (api *ApiHandler) exchangeCodeForToken(code string) (types.InstagramToken, error) {
    const tokenURL = "https://api.instagram.com/oauth/access_token"
    
    if api.clientID == "" || api.clientSecret == "" || api.redirect_uri== "" {
      return types.InstagramToken{}, fmt.Errorf("missing required credentials (client_id, client_secret, redirect_uri)")

    }
    
    form := url.Values{}
    form.Add("client_id", api.clientID)
    form.Add("client_secret", api.clientSecret)
    form.Add("grant_type", "authorization_code")
    form.Add("code", code)
    form.Add("redirect_uri", api.redirect_uri) // You might be missing this!

    req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
    if err != nil {
        return types.InstagramToken{}, fmt.Errorf("creating request: %w", err)
    }
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    resp, err := api.httpClient.Do(req)
    if err != nil {
        return types.InstagramToken{}, fmt.Errorf("executing request: %w", err)
    }
    defer resp.Body.Close()

    // Read and log the response body for debugging
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return types.InstagramToken{}, fmt.Errorf("reading response body: %w", err)
    }
    log.Printf("Instagram API Response: Status=%d, Body=%s", resp.StatusCode, string(body))

    if resp.StatusCode != http.StatusOK {
        return types.InstagramToken{}, fmt.Errorf("instagram API returned status %d: %s", resp.StatusCode, string(body))
    }

    // Create a new reader with the body we read
    var token types.InstagramToken
    if err := json.NewDecoder(strings.NewReader(string(body))).Decode(&token); err != nil {
        return types.InstagramToken{}, fmt.Errorf("decoding response: %w", err)
    }

    return token, nil
}


func (api *ApiHandler) StopMonitoring() {
	api.monitorCancel()
}
