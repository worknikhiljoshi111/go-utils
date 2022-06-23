package aws_test

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/credifranco/stori-utils-go/aws"
	"github.com/stretchr/testify/assert"
)

func TestParseS3URL(t *testing.T) {
	urls := [3]string{
		"https://s3.amazonaws.com/stori-bucket/stori-object",
		"https://stori-bucket.s3.amazonaws.com/stori-object",
		"s3://stori-bucket/stori-object",
	}
	for _, url := range urls {
		parsedUrlObject, err := aws.ParseS3URL(url)

		if err != nil {
			t.Error("parsing error", err)
		}

		if parsedUrlObject.BucketName != "stori-bucket" {
			t.Errorf("expected 'stori-bucket', got '%s'", parsedUrlObject.BucketName)
		}
		if parsedUrlObject.Key != "stori-object" {
			t.Errorf("expected 'stori-object', got '%s'", parsedUrlObject.Key)
		}
		if parsedUrlObject.Url != url {
			t.Errorf("expected '%s', got '%s'", url, parsedUrlObject.Url)
		}
	}
}

func TestNewLambdaClient(t *testing.T) {
	// temporarily clear the AWS_REGION env variable
	key := "AWS_REGION"
	val := os.Getenv(key)

	os.Unsetenv(key)
	t.Cleanup(func() {
		os.Setenv(key, val)
	})

	_, err := aws.NewLambdaClient()

	assert.ErrorIs(
		t,
		err,
		aws.ErrRegionNotSet,
		"should get error back if AWS_REGION env variable is not set",
	)
}

func TestNewSNSClient(t *testing.T) {
	SNSClient, err := aws.NewSNSClient()
	assert.NoError(t, err, "there should not be an unexpected error")
	assert.NotNil(t, SNSClient, "SNSClient should not be null")
}

// Lambda client used to avoid hitting actual AWS endpoint when unit tests run on the GitHub
type mockLambdaClient struct {
	lambdaiface.LambdaAPI
}

// Perform no action and return a dummy revision ID
func (m *mockLambdaClient) Invoke(*lambda.InvokeInput) (*lambda.InvokeOutput, error) {
	statusCode := int64(http.StatusOK)
	payload := "hola 2021"
	return &lambda.InvokeOutput{StatusCode: &statusCode, Payload: []byte(payload)}, nil
}

func TestInvokeLambda(t *testing.T) {
	// Action
	li := aws.LambdaInvocation{}
	mockSvc := &mockLambdaClient{}

	resp, err := li.InvokeLambda(mockSvc)
	assert.NoError(t, err)

	assert.Equal(t, *resp.StatusCode, int64(http.StatusOK), "should not be able to make the request successful if you don't have the lambda info correctly")

	assert.Equal(t, resp.Payload, []byte("hola 2021"), "should not be able to read the payload information if the request fails")
}

type mockS3DownloadClient struct {
	s3manageriface.DownloaderAPI
	op int64
}

// Return Dummy Output
func (m *mockS3DownloadClient) Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (int64, error) {
	return m.op, nil
}
func TestDownload(t *testing.T) {

	s3FileURL := "https://test-nikbucket.s3.us-west-2.amazonaws.com/test/nikhil/testfile"
	s3Obj, _ := aws.ParseS3URL(s3FileURL)
	s3Obj.LocalPath = "testfile.txt"

	mocks3 := &mockS3DownloadClient{op: 123123921}

	resp, err := s3Obj.Download(mocks3)

	assert.NoError(t, err)
	assert.Equal(t, resp, int64(123123921), "should not be able to read resp if download fails")
}

type mocks3UploadClient struct {
	s3manageriface.UploaderAPI
	s3manager.UploadOutput
}

// Returns dummy UploadOutput
func (m *mocks3UploadClient) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {

	return &m.UploadOutput, nil
}

func TestUpload(t *testing.T) {
	s3FileURL := "https://test-nikbucket.s3.us-west-2.amazonaws.com/test/nikhil/testfile"
	s3Obj, _ := aws.ParseS3URL(s3FileURL)
	s3Obj.LocalPath = "testfile.txt"

	versionId := "1.0.0"
	etag := "123421"
	mocks3 := &mocks3UploadClient{UploadOutput: s3manager.UploadOutput{Location: "location", UploadID: "12334", VersionID: &versionId, ETag: &etag}}
	resp, err := s3Obj.Upload(mocks3)
	assert.NoError(t, err)
	assert.Equal(t, resp.UploadID, "12334", "should not be able to read the uploadid if upload fails")
	assert.Equal(t, resp.Location, "location", "should not be able to read the location if upload fails")
	assert.Equal(t, *resp.VersionID, "1.0.0", "should not be able to read the version if upload fails")
	assert.Equal(t, *resp.ETag, "123421", "should not be able to read the Etag if upload fails")

}
