package types

import (
	"time"
)


type InstagramToken struct {
  AccessToken string `json:"access_token" dynamodbav:"access_token"`
  TokenType string `json:"token_type" dynamodbav:"token_type"`
  ExpiresAt time.Time `json:"created_at" dynamodbav:"created_at"`
}


type InstagramMedia struct{
  ID string `json:"id"`
  Caption string `json:"caption"`
  MediaURL string `json:"media_url"`
  Owner Owner `json:"owner"`
  Timestamp string `json:"timestamp"`
}
type Owner struct{
  ID string `json:"id"`
}
type MediaResponse struct{
  Data []InstagramMedia `json:"media"`
}


