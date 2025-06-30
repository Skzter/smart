package main

import (
	"context"
	"flag"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func main() {
	bucketName := flag.String("bucket", "smart-autotester", "S3 bucket name")
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"))
	if err != nil {
		slog.Error("unable to load SDK config", "error", err)
		return
	}

	s3Client := s3.NewFromConfig(cfg)
	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(*bucketName),
	})

	if err != nil {
		slog.Info("bucket does not exist, trying to create new bucket")
		_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(*bucketName),
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraintEuCentral1,
			},
		})

		if err != nil {
			slog.Error("failed to create bucket", "error", err)
			return
		}

		slog.Info("Successfully created bucket", "bucket", bucketName)
	}

	_, err = s3Client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(*bucketName),
		PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	})

	if err != nil {
		slog.Error("error setting public access block", "error", err)
		return
	}

	_, err = s3Client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(*bucketName),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	})

	if err != nil {
		slog.Error("error setting versioning status", "error", err)
		return
	}

	slog.Info("Successfully set versioning and security option")
}
