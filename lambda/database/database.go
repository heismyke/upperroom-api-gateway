package database

import (
	"lambda-func/types"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type AuthStore interface{
  StoreToken(token types.InstagramToken) error
}

const (
  TABLE_NAME = "AuthTablee"
)


type DynamoDBClient struct{
  databaseStore *dynamodb.DynamoDB
}

func NewDynamoDBClient() *DynamoDBClient{
  dbSession := session.Must(session.NewSession())
  db := dynamodb.New(dbSession)
  return &DynamoDBClient{
    databaseStore: db,
  }
}

func (db *DynamoDBClient) StoreToken(token types.InstagramToken) error {
  item := &dynamodb.PutItemInput{
    TableName : aws.String(TABLE_NAME),
    Item : map[string]*dynamodb.AttributeValue{
      "access_token" : {
        S : aws.String(token.AccessToken), 
      },
      "token_type" : {
        S: aws.String(token.TokenType),
      },
      "expires_at": {
        S : aws.String(token.ExpiresAt.Format(time.RFC3339)),
      },
    },
  }
  _, err := db.databaseStore.PutItem(item)
  if err != nil {
    return err
  }
  return nil
}

