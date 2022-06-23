package user_test

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/credifranco/stori-utils-go/user"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestGetCognitoUser will test retrieval of Cognito attributes using mock data
func TestGetCognitoUser(t *testing.T) {

	// testing for all the values
	claims := make(map[string]interface{})
	claims["sub"] = "da3ddc9a-ab76-4f6e-a1f4-9f6dc4e3095f"
	claims["email"] = "stori@storicard.com"
	claims["phone_number"] = "+527772536437"

	authorizer := make(map[string]interface{})
	authorizer["claims"] = claims

	e := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: authorizer,
		},
	}

	cognitoUser, err := user.GetCognitoUser(e.RequestContext)
	assert.NoError(t, err, "should not be able to get an error if the payload is valid")
	assert.Equal(t, "da3ddc9a-ab76-4f6e-a1f4-9f6dc4e3095f", cognitoUser.Sub, "should be the expected value if the payload is valid")
	assert.Equal(t, "stori@storicard.com", cognitoUser.Email, "should be the expected value if the payload is valid")
	assert.Equal(t, "+527772536437", cognitoUser.PhoneNumber, "should be the expected value if the payload is valid")

	// invalid user id - uuid not valid
	claims = make(map[string]interface{})
	claims["sub"] = "36c5ba10-0000-0000-0000-jtsyhdtstyhw"
	authorizer = make(map[string]interface{})
	authorizer["claims"] = claims

	e = events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: authorizer,
		},
	}

	cognitoUser, err = user.GetCognitoUser(e.RequestContext)
	assert.EqualError(t, err, "sub: must be a valid UUID v4.", "should be able to get an error because the field's userid is invalid")

	// testing for just the sub
	claims = make(map[string]interface{})
	claims["sub"] = "da3ddc9a-ab76-4f6e-a1f4-9f6dc4e3095f"
	authorizer = make(map[string]interface{})
	authorizer["claims"] = claims

	e = events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: authorizer,
		},
	}

	cognitoUser, err = user.GetCognitoUser(e.RequestContext)
	assert.NoError(t, err, "should not be able to get an error if the payload is valid")
	assert.Equal(t, "da3ddc9a-ab76-4f6e-a1f4-9f6dc4e3095f", cognitoUser.Sub, "should be the expected value if the payload is valid")
	assert.Equal(t, "", cognitoUser.Email, "should be the expected value if the payload is valid")

}
