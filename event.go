package consumer

import "github.com/aws/aws-sdk-go/service/sqs"

const (
	EventReceiveMessage      = "ReceiveMessage"
	EventProcessMessage      = "ProcessMessage"
	EventReceiveMessageError = "ReceiveMessageError"
)

type OnReceiveMessage func(messages []*sqs.Message)
type OnProcessMessage func(message *sqs.Message)
type OnReceiveMessageError func(err error)
