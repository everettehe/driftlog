package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Resource represents a live AWS resource with its type, ID, and attributes.
type Resource struct {
	Type       string
	ID         string
	Attributes map[string]string
}

// Fetcher retrieves live AWS resource state.
type Fetcher struct {
	ec2Client *ec2.Client
	s3Client  *s3.Client
}

// NewFetcher creates a Fetcher using the default AWS config and region.
func NewFetcher(ctx context.Context, region string) (*Fetcher, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}
	return &Fetcher{
		ec2Client: ec2.NewFromConfig(cfg),
		s3Client:  s3.NewFromConfig(cfg),
	}, nil
}

// FetchEC2Instances returns live EC2 instance resources.
func (f *Fetcher) FetchEC2Instances(ctx context.Context) ([]Resource, error) {
	out, err := f.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("describing ec2 instances: %w", err)
	}

	var resources []Resource
	for _, reservation := range out.Reservations {
		for _, inst := range reservation.Instances {
			attrs := map[string]string{
				"instance_type": string(inst.InstanceType),
				"state":         string(inst.State.Name),
			}
			if inst.ImageId != nil {
				attrs["ami"] = aws.ToString(inst.ImageId)
			}
			resources = append(resources, Resource{
				Type:       "aws_instance",
				ID:         aws.ToString(inst.InstanceId),
				Attributes: attrs,
			})
		}
	}
	return resources, nil
}

// FetchS3Buckets returns live S3 bucket resources.
func (f *Fetcher) FetchS3Buckets(ctx context.Context) ([]Resource, error) {
	out, err := f.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("listing s3 buckets: %w", err)
	}

	var resources []Resource
	for _, b := range out.Buckets {
		resources = append(resources, Resource{
			Type:       "aws_s3_bucket",
			ID:         aws.ToString(b.Name),
			Attributes: map[string]string{"bucket": aws.ToString(b.Name)},
		})
	}
	return resources, nil
}
