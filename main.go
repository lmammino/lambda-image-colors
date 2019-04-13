package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/lmammino/lambda-image-colors/utils"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type ImageRecord struct {
	ID     string   `json:"-"`
	Key    string   `json:"key"`
	Bucket string   `json:"bucket"`
	Colors []string `json:"colors"`
}

func HandleRequest(ctx context.Context, event events.S3Event) error {
	palette := utils.Palette{
		"red":       {255, 0, 0},
		"orange":    {255, 165, 0},
		"yellow":    {255, 255, 0},
		"green":     {0, 255, 0},
		"turquoise": {0, 222, 222},
		"blue":      {0, 0, 255},
		"violet":    {128, 0, 255},
		"pink":      {255, 0, 255},
		"brown":     {160, 82, 45},
		"black":     {0, 0, 0},
		"gray":      {128, 128, 128},
		"white":     {255, 255, 255},
	}

	awsSession := session.New()
	s3Client := s3.New(awsSession)

	esHosts := os.Getenv("ES_HOSTS")
	if esHosts == "" {
		esHosts = "http://localhost:9200"
	}
	esIndex := os.Getenv("ES_INDEX")
	if esIndex == "" {
		esIndex = "images"
	}

	cfg := elasticsearch.Config{
		Addresses: strings.Split(esHosts, ","),
	}
	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return err
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

		colors, err := utils.GetProminentColors(s3File.Body, palette)

		if err != nil {
			return err
		}

		newRecord := ImageRecord{
			ID:     "s3://" + bucket + "/" + key,
			Bucket: bucket,
			Key:    key,
			Colors: colors,
		}

		// index the record in elastic search
		body, err := json.Marshal(newRecord)
		if err != nil {
			return err
		}

		req := esapi.IndexRequest{
			Index:      esIndex,
			DocumentID: newRecord.ID,
			Body:       bytes.NewReader(body),
			Refresh:    "true",
		}

		// Perform the request with the client.
		res, err := req.Do(context.Background(), esClient)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.IsError() {
			return fmt.Errorf("[%s] Error indexing document ID=%s", res.Status(), newRecord.ID)
		}
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
