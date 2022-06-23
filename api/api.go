package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/credifranco/stori-utils-go/aws"
	"github.com/credifranco/stori-utils-go/db"
	slog "github.com/credifranco/stori-utils-go/log"
	"github.com/credifranco/stori-utils-go/redis"
	"go.uber.org/zap"
)

type StoriServicesConfig struct {
	DBProxy    *db.DBProxies
	Lambda     bool
	RedisCache bool // TODO
	SNS        bool // TODO
}

type StoriServices struct {
	DB        db.DBConnector
	LambdaAPI lambdaiface.LambdaAPI
	Logger    *zap.SugaredLogger
	SNSAPI    snsiface.SNSAPI
	Redis     *redis.RedisConn
}

type APIGatewayHandlerFunc func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type errorResponseDetails struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type errorResponse struct {
	Error errorResponseDetails `json:"error"`
}

const invalidJSONError = "could not create JSON response"

// NewStoriServices is a constructor for StoriServices. The Logger field of the StoriServices will
// always be created. DB, LA & SNS are optional, set based on the values passed in via the StoriServicesConfig.
func NewStoriServices(ctx context.Context, config StoriServicesConfig) (StoriServices, error) {
	logger, err := slog.NewLogger()
	if err != nil {
		return StoriServices{}, fmt.Errorf("error creating logger: %v", err)
	}

	out := StoriServices{Logger: logger}

	if config.DBProxy != nil {
		a := &db.AWSDB{}
		if err := a.NewConnection(ctx, *config.DBProxy); err != nil {
			return StoriServices{}, fmt.Errorf("error establishing database connection: %v", err)
		}
		out.DB = a
	}

	if config.Lambda {
		if out.LambdaAPI, err = aws.NewLambdaClient(); err != nil {
			return StoriServices{}, fmt.Errorf("error creating lambda client: %v", err)
		}
	}
	// Create SNS service client
	if config.SNS {
		if out.SNSAPI, err = aws.NewSNSClient(); err != nil {
			return StoriServices{}, fmt.Errorf("error creating SNS client: %v", err)
		}
	}

	if config.RedisCache {
		r := redis.RedisConn{}
		if err = r.NewConn(ctx); err != nil {
			return StoriServices{}, fmt.Errorf("error creating Redis client: %v", err)
		}
		out.Redis = r

	}
	return out, nil
}

// JSONResponse creates a APIGatewayProxyResponse with properly formed headers and body. If body
// cannont be parsed into JSON a code 500 will be returned in the APIGatewayProxyResponse.
func JSONResponse(code int, body interface{}) (events.APIGatewayProxyResponse, error) {
	bb, err := json.Marshal(body)
	if err != nil {
		// we can recover to respond with something nice
		return JSONErrResponse(http.StatusInternalServerError, invalidJSONError)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       string(bb),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// JSONErrResponse creates a APIGatewayProxyResponse with properly formed headers and body. code is
// expected to be a valid value for http.StatusText().
func JSONErrResponse(code int, message string) (events.APIGatewayProxyResponse, error) {
	body := errorResponse{Error: errorResponseDetails{code, message, http.StatusText(code)}}
	bb, err := json.Marshal(body)
	if err != nil {
		// we've reached the end of where we can recover with something nice programatically.
		// So let's hand write some JSON
		code := http.StatusInternalServerError
		message := invalidJSONError
		status := http.StatusText(code)
		bodyText := fmt.Sprintf(
			`{"error":{"code":%d,"message":"%s","status":"%s"}}`,
			code,
			message,
			status,
		)

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       bodyText,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       string(bb),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}
