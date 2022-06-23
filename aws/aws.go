package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/joho/godotenv"
)

type S3Interface interface {
	Download() (int64, error)
	Upload() (int64, error)
	NewConn() error
}

// S3Object represents the components of an S3 URL
type S3Object struct {
	BucketName string
	Key        string
	Url        string
	LocalPath  string
}

// Secret contains information about a secret stored in AWS Secret Manager.
type secret struct {
	id string
}

// LambdaInvocation contains information needed to create a Lambda invocation.
type LambdaInvocation struct {
	FunctionName   string
	Event          interface{}
	InvocationType string
}

var awsConfig aws.Config
var ErrRegionNotSet = errors.New("aws region is not set")

func init() {
	// godotenv.Load sets values found in .env file, but does not override existing env values
	_ = godotenv.Load()

	if region, ok := os.LookupEnv("AWS_REGION"); !ok {
		log.Println(ErrRegionNotSet)
	} else {
		awsConfig = aws.Config{Region: aws.String(region)}
	}
}

// ParseS3URL parses an S3 url and returns the bucket name, path to the object and the url in an S3Object struct.
// The following types of S3 urls can be parsed:
// - https://s3.amazonaws.com/[bucket_name]/[path_to_object]
// - https://[bucket_name].s3.amazonaws.com/[path_to_object]
// - s3://[bucket_name]/[path_to_object]
// If the url doesn't match any of these pattern then a nil along with an error is returned.
func ParseS3URL(s3URL string) (S3Object, error) {

	u, err := url.Parse(s3URL)
	if err != nil {
		return S3Object{}, errors.New("couldn't parse the url")
	}

	if u.Scheme != "s3" && u.Scheme != "https" {
		return S3Object{}, errors.New("invalid protocol")
	}

	parsedObject := S3Object{
		BucketName: u.Host,
		Key:        u.Path[1:],
		Url:        s3URL,
	}

	if strings.Contains(s3URL, "https://s3") {
		stringMap := strings.Split(u.Path, "/")
		parsedObject.BucketName = stringMap[1]
		parsedObject.Key = strings.Join(stringMap[2:], "/")
	} else {
		parsedObject.BucketName = strings.Split(u.Host, ".s3")[0]
	}

	return parsedObject, nil
}

// GetRDSAuthToken uses the given RDS credentials to create an IAM authentication token
func GetRDSAuthToken(ctx context.Context, host, user string, port int) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}

	return auth.BuildAuthToken(
		ctx,
		fmt.Sprintf("%s:%d", host, port),
		*awsConfig.Region,
		user,
		cfg.Credentials,
	)
}

// GetSecret returns the secret from AWS Secrets Manager as a JSON-encoded string.
// If the secret does not exist, an error is returned.
func GetSecret(id string) (string, error) {
	s := secret{id: id}

	sess := session.Must(session.NewSession(&awsConfig))
	svc := secretsmanager.New(sess)

	return s.getSecretString(svc)
}

// getSecretString fetches the secret from AWS Secrets Manager using a
// secretsmanageriface.SecretsManagerAPI interface
func (s secret) getSecretString(sma secretsmanageriface.SecretsManagerAPI) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(s.id),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := sma.GetSecretValue(input)
	if err != nil {
		return "", errors.New("secret not found")
	}

	return *result.SecretString, nil
}

// InvokeLambda invokes a lambda with an event defined in the LambdaInvocation struct. This event
// can be any arbitray struct.
func (li LambdaInvocation) InvokeLambda(la lambdaiface.LambdaAPI) (*lambda.InvokeOutput, error) {

	payload, err := json.Marshal(li.Event)
	if err != nil {
		return &lambda.InvokeOutput{}, errors.New("error parsing json from lambda event")
	}

	return la.Invoke(
		&lambda.InvokeInput{
			FunctionName:   &li.FunctionName,
			Payload:        payload,
			InvocationType: &li.InvocationType,
		},
	)
}

// NewLambdaClient is a helper function to create a new lambda client with default config
func NewLambdaClient() (*lambda.Lambda, error) {
	region, ok := os.LookupEnv("AWS_REGION")
	if !ok {
		return nil, ErrRegionNotSet
	}

	sess, err := session.NewSession(&aws.Config{
		Region: &region,
	})
	if err != nil {
		return nil, err
	}

	return lambda.New(sess), nil
}

// NewSNSClient is a helper function to create a new SNS client with default config
func NewSNSClient() (*sns.SNS, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return sns.New(sess), nil
}

// Upload will upload a single file to S3,
// it will require a pre-built aws session
// return error if the file uploading has failed
func (s S3Object) Upload(u s3manageriface.UploaderAPI) (*s3manager.UploadOutput, error) {

	// Open the file for use
	file, err := os.Open(s.LocalPath)

	if err != nil {
		return &s3manager.UploadOutput{}, errors.New("error opening file")
	}
	defer file.Close()
	uploadOutput, err := u.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(s.Key),
		Body:   file,
	})

	if err != nil {
		return &s3manager.UploadOutput{}, errors.New("error in uploading file")
	}
	return uploadOutput, nil
}

// Download will download single file from s3
// it will require a pre-built aws session
// return int64 if successful  and error if the file downloading has failed
func (s S3Object) Download(d s3manageriface.DownloaderAPI) (int64, error) {

	// Open the file for use
	file, err := os.Create(s.LocalPath)

	if err != nil {
		return 0, errors.New("error opening file")

	}
	defer file.Close()
	downloadOutput, err := d.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(s.BucketName),
			Key:    aws.String(s.Key),
		})
	if err != nil {
		return 0, errors.New("error in download file")
	}
	return downloadOutput, nil
}
