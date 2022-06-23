package api_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"

	"github.com/credifranco/stori-utils-go/api"
)

func TestNewStoriServices(t *testing.T) {
	handler, err := api.NewStoriServices(context.Background(), api.StoriServicesConfig{})

	assert.NoError(t, err, "there should not be an unexpected error")
	assert.Nil(t, handler.DB, "handler with zero config should have a nil DB")
	assert.Nil(t, handler.LambdaAPI, "handler with zero config should have a nil LA")
	assert.Nil(t, handler.SNSAPI, "handler with zero config should have a nil SNS")
	assert.Nil(t, handler.Redis, "handler with zero config should have a nil Redis")
	assert.NotNil(t, handler.Logger, "handler with zero config should have a non-nil logger")
}

func TestJSONResponseErr(t *testing.T) {
	// Setup
	a := assert.New(t)

	testCode := 200
	testBody := map[string]interface{}{
		"foo": make(chan (int)),
	}

	// Action
	resp, err := api.JSONResponse(testCode, testBody)

	// Assertions
	a.Equal(resp.StatusCode, http.StatusInternalServerError)
	a.Nil(err)
}

func TestJSONResponseNoErr(t *testing.T) {
	// Setup
	a := assert.New(t)

	testCode := 200
	testBody := map[string]interface{}{
		"foo": "bar",
	}

	// Action
	resp, err := api.JSONResponse(testCode, testBody)

	// Assertions
	a.Equal(resp.StatusCode, http.StatusOK)
	a.Equal(resp.Body, `{"foo":"bar"}`)
	a.Nil(err)
}

func TestInvalidJSON(t *testing.T) {
	a := assert.New(t)
	// make something that will cause json.Marshal to return an error
	type cyclic struct {
		ID   int     `json:"id"`
		Next *cyclic `json:"next"`
	}

	c1 := cyclic{1, nil}
	c2 := cyclic{1, &c1}
	c1.Next = &c2

	expected := events.APIGatewayProxyResponse{
		Body:       `{"error":{"code":500,"message":"could not create JSON response","status":"Internal Server Error"}}`,
		Headers:    map[string]string{"Content-Type": "application/json"},
		StatusCode: 500,
	}

	res, _ := api.JSONResponse(200, c1)

	a.Equal(
		res,
		expected,
		"invalid JSON object should create a 500 error",
	)
}

func TestJSONErrResponse(t *testing.T) {
	a := assert.New(t)

	expected := events.APIGatewayProxyResponse{
		Body:       `{"error":{"code":400,"message":"bad news","status":"Bad Request"}}`,
		Headers:    map[string]string{"Content-Type": "application/json"},
		StatusCode: 400,
	}

	res, _ := api.JSONErrResponse(http.StatusBadRequest, "bad news")

	a.Equal(
		res,
		expected,
		"body text should be properly formatted JSON and status should be correct HTTP Status Code text",
	)

	expected = events.APIGatewayProxyResponse{
		Body:       `{"error":{"code":42,"message":"","status":""}}`,
		Headers:    map[string]string{"Content-Type": "application/json"},
		StatusCode: 42,
	}

	res, _ = api.JSONErrResponse(42, "")

	a.Equal(
		res,
		expected,
		"status text should be empty string if code is not valid HTTP Status Code",
	)
}
