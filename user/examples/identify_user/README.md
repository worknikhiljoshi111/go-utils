# Identify the user by id

## Example usage

The example [AWS SAM Template](/user/examples/identify_user/template.yaml) is a minimal template to allow the invocation of the example Lambda using  `sam local start-api`

```sh
# build binary
cd lambda && GOARCH=amd64 GOOS=linux go build -o main && cd ..

# start local SAM server
sam local start-api

# make API request with JSON file that contains event data
sam local invoke IdentifyUserFunction --event event.json
```

expected result
```json
{
  "statusCode": 200,
  "headers": {
    "Content-Type": "application/json"
  },
  "multiValueHeaders": null,
  "body": "Hello c6cabfe7-82cb-4d73-9993-a1783c224e19"
}
```