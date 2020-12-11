package consumer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func NewClient(region *string) *sqs.SQS {
	sess, err := session.NewSession(&aws.Config{
		Region: region},
	)

	if err != nil {
		panic(err)
	}

	return sqs.New(sess)
}
