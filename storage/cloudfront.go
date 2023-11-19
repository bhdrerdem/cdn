package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/rs/zerolog/log"
)

type CloudFrontService struct {
	DistributionID  string
	DistributionUrl string
	Client          *cloudfront.Client
}

func NewCloudFrontService(distributionID, distributionURL string) *CloudFrontService {

	if distributionID == "" || distributionURL == "" {
		log.Fatal().Msg("Distribution id or url cannot be empty.")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load aws config")
	}

	client := cloudfront.NewFromConfig(cfg)

	return &CloudFrontService{
		DistributionID:  distributionID,
		DistributionUrl: distributionURL,
		Client:          client,
	}
}

func (c *CloudFrontService) CreateInvalidation(path string) error {

	callerReference := aws.String(time.Now().Format(time.RFC3339Nano))

	invalidationInput := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(c.DistributionID),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: callerReference,
			Paths: &types.Paths{
				Quantity: aws.Int32(1),
				Items:    []string{path},
			},
		},
	}

	_, err := c.Client.CreateInvalidation(context.Background(), invalidationInput)
	if err != nil {
		return err
	}

	return nil
}

func (c *CloudFrontService) FetchFile(key string) (io.ReadCloser, int64, string, error) {

	response, err := http.Get(fmt.Sprintf("%s/%s", c.DistributionUrl, key))
	if err != nil || response.StatusCode != http.StatusOK {
		return nil, 0, "", err
	}

	contentType := response.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return response.Body, response.ContentLength, contentType, nil
}
