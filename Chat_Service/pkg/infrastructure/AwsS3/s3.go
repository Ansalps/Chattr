package awss3

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AwsConfig struct {
	AwsAccess      string
	AwsSecretAcces string
	AwsRegion      string
}

func NewS3Client(awsAccess, awsSecretAccess, awsRegion string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				awsAccess,
				awsSecretAccess,
				"",
			),
		),
	)
	if err != nil {
		log.Println("print the error in NewS3Client", err)
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg)
	//fmt.Println("s3Client", s3Client)
	return s3Client, nil
}
