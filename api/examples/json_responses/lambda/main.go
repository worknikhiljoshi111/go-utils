package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/credifranco/stori-utils-go/api"
)

func handler(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// error
	if val := e.QueryStringParameters["error"]; val == "true" {
		return api.JSONErrResponse(http.StatusInternalServerError, "something bad happened")
	}

	// success
	out := struct {
		Name string `json:"name"`
		Id   int    `json:"id"`
	}{
		Name: "William Shakespeare",
		Id:   1,
	}

	return api.JSONResponse(http.StatusOK, out)
}

func main() {
	lambda.Start(handler)
}
