package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/credifranco/stori-utils-go/api"
	"github.com/credifranco/stori-utils-go/user"
)

func handler(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cognitoUser, err := user.GetCognitoUser(e.RequestContext)
	if err != nil {
		return api.JSONErrResponse(http.StatusUnauthorized, "upss, unauthorized")
	}

	out := fmt.Sprintf("Hello %v", cognitoUser.Sub)

	return api.JSONResponse(http.StatusOK, out)
}

func main() {
	lambda.Start(handler)
}
