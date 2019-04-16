package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/lmammino/lambda-image-colors/cmd/utils"

	elasticsearch "github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
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

	esHosts := os.Getenv("ELASTIC_HOSTS")
	if esHosts == "" {
		esHosts = "http://localhost:9200"
	}
	esIndex := os.Getenv("ELASTIC_INDEX")
	if esIndex == "" {
		esIndex = "images"
	}

	cfg := elasticsearch.Config{
		Addresses: strings.Split(esHosts, ","),
		Transport: &http.Transport{
			ResponseHeaderTimeout: 5 * time.Second,
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		},
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

		colors, err := utils.GetProminentColors(s3File.Body, *palette)

		if err != nil {
			return err
		}

		newRecord := ImageRecord{
			ID:     bucket + "|" + key,
			Bucket: bucket,
			Key:    key,
			Colors: colors,
		}

		// index the record in elastic search
		body, err := json.Marshal(newRecord)
		if err != nil {
			return err
		}

		fmt.Printf("Indexing %s with colors %v\n", newRecord.ID, newRecord.Colors)

		req := esapi.IndexRequest{
			Index:        esIndex,
			DocumentType: "default",
			DocumentID:   newRecord.ID,
			Body:         bytes.NewReader(body),
			Refresh:      "true",
		}

		// Perform the request with the client.
		res, err := req.Do(context.Background(), esClient)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.IsError() {
			return fmt.Errorf("[%s] Error indexing document ID=%s:/n%s", res.Status(), newRecord.ID, res.String())
		}
	}

	return nil
}

func main() {
	fmt.Println("Starting Lambda ...")
	lambda.Start(HandleRequest)
}
