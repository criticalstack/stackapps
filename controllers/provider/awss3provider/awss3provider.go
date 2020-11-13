package awss3provider

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	"github.com/pkg/errors"
)

var Type featuresv1alpha1.StackValueSourceType = featuresv1alpha1.StackValueSourceAWSS3

type AWSS3Provider struct {
	config *featuresv1alpha1.StackValueSource
	key    string
}

func (p AWSS3Provider) Values() (interface{}, error) {
	if p.config.Region == "" {
		return nil, errors.New("Please add Region to awss3 config in the stackapps config for this namespace")
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(p.config.Region),
	}))
	downloader := s3manager.NewDownloader(sess)
	var buffer aws.WriteAtBuffer
	// Write the contents of S3 Object to the file
	_, err := downloader.Download(&buffer, &s3.GetObjectInput{
		Bucket: aws.String(p.config.Route),
		Key:    aws.String(p.key),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to download file from S3")
	}
	return buffer.Bytes(), nil
}

func New(c *featuresv1alpha1.StackValueSource, path string) AWSS3Provider {
	return AWSS3Provider{
		config: c,
		key:    path,
	}
}
