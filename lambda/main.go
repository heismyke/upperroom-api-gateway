package main

import (
	"lambda-func/app"

	"github.com/aws/aws-lambda-go/lambda"
)


func main() {
  myApp := app.New()
  lambda.Start(myApp.ApiHandler.StoreTokenHandler) 
}
