# StoriServices

**StoriServices** is a struct that carries common utilities used in our AWS Lambdas and other services. **NewStoriServices** is a constructor function that eliminates much of the boilerplate code needed to create the underlying database, lambda client, and logger services.

## Example Usage

[lambda/main.go](/api/examples/stori_services/lambda/main.go) shows a common implementation of StoriServices in an AWS lambda.

The example [AWS SAM Template](/api/examples/stori_services/template.yaml) is a minimal template to allow the invocation of the example Lambda using  `sam local start-api`

```sh
# build binary
cd lambda && GOARCH=amd64 GOOS=linux go build -o main && cd ..

# start local SAM server
sam local start-api

# make API request
curl http://localhost:3000/services

# expected result
{"message":"hello world!"}
```
