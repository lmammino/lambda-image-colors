package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/lmammino/lambda-image-colors/cmd/utils"
)

type ImageRecord struct {
	ID     string   `json:"-"`
	Key    string   `json:"key"`
	Bucket string   `json:"bucket"`
	Colors []string `json:"colors"`
}

func HandleRequest(ctx context.Context, event events.S3Event) error {
	palette := utils.GetDefaultPalette()

	s3Endpoint := os.Getenv("AWS_S3_ENDPOINT")
	var s3ApiUsePathEndpoint bool
	if len(os.Getenv("AWS_S3_USE_PATH")) > 0 {
		s3ApiUsePathEndpoint = true
	}

	awsSession := session.New()
	s3Client := s3.New(awsSession)
	if len(s3Endpoint) > 0 {
		s3Client = s3.New(
			awsSession,
			&aws.Config{
				S3ForcePathStyle: &s3ApiUsePathEndpoint,
				Endpoint:         &s3Endpoint,
			},
		)
	}

	for _, s3record := range event.Records {
		bucket := s3record.S3.Bucket.Name
		key := s3record.S3.Object.Key

		s3GetObjectInput := &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &key,
		}
		s3File, err := s3Client.GetObject(s3GetObjectInput)
		if err != nil {
			return err
		}

		colors, err := utils.GetProminentColors(s3File.Body, *palette)

		if err != nil {
			return err
		}

		fmt.Printf("Indexing s3://%s/%s with colors -> %v\n", bucket, key, colors)

		tags := []*s3.Tag{}
		for i, color := range colors {
			tagKey := fmt.Sprintf("Color%d", i+1)
			tag := s3.Tag{Key: aws.String(tagKey), Value: aws.String(color)}
			tags = append(tags, &tag)
		}

		taggingRequest := &s3.PutObjectTaggingInput{
			Bucket: &bucket,
			Key:    &key,
			Tagging: &s3.Tagging{
				TagSet: tags,
			},
		}

		_, err = s3Client.PutObjectTagging(taggingRequest)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	fmt.Println("Starting Lambda ...")
	lambda.Start(HandleRequest)
}
