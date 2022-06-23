# API JSON Responses

**JSONResponse**  and **JSONERRResponse** are both helper functions to return properly formatted JSON-encoded AWS APIGatewayProxyResponse from Lambda functions.

## Example Usage

The example [AWS SAM Template](/api/examples/json_responses/template.yaml) is a minimal template to allow the invocation of the example Lambda using  `sam local start-api`

```sh
# start local SAM server
sam local start-api

# make API request

# success
curl http://localhost:3000/json

# expected result
{"name":"William Shakespeare","id":1}

# error 
curl http://localhost:3000/json\?error\=true

# expected result
{"error":{"code":500,"message":"something bad happened","status":"Internal Server Error"}}
```
