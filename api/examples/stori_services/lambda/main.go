package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/credifranco/stori-utils-go/api"
	"github.com/credifranco/stori-utils-go/aws"
	"github.com/credifranco/stori-utils-go/db"
)

// handler wraps an api.APIGatewayHandlerFunc, allowing access to passed in StoriServices
func handler(s api.StoriServices) api.APIGatewayHandlerFunc {
	return func(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		if val := e.QueryStringParameters["error"]; val == "true" {
			// use the logger from the passed in services
			s.Logger.Errorw("Invalid params", "params", e.QueryStringParameters)
			return api.JSONErrResponse(http.StatusBadRequest, "invalid parameters")
		}

		// use the passed in db interface to get some data
		var msg string
		if err := s.DB.QueryRow(ctx, "SELECT 'hello world!'").Scan(&msg); err != nil {
			return api.JSONErrResponse(http.StatusInternalServerError, "database error: "+err.Error())
		}
		s.Logger.Infow("DB Result", "msg", msg)

		// use the lambda API
		type lambdaEvent struct {
			Body map[string]string `json:"body"`
		}
		li := aws.LambdaInvocation{
			FunctionName:   "some-lambda",
			Event:          lambdaEvent{map[string]string{"key": "value"}},
			InvocationType: "RequestResponse",
		}
		if res, err := li.InvokeLambda(s.LambdaAPI); err != nil {
			// do nothing for this example, this isn't calling a real lambda
			s.Logger.Errorw("Lambda error", "err", err, "resposne", res)
		}

		return api.JSONResponse(
			http.StatusOK,
			struct {
				Message string `json:"message"`
			}{msg},
		)
	}
}

func main() {
	// get a StoriServices struct with a DB reader proxy
	proxy := db.Read
	s, err := api.NewStoriServices(
		context.Background(),
		api.StoriServicesConfig{DBProxy: &proxy, Lambda: true},
	)
	if err != nil {
		log.Fatalf("error creating StoriServices: %v", err)
	}

	// call the lambda
	lambda.Start(handler(s))
}
