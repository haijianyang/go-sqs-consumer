package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	consumer "github.com/haijianyang/go-sqs-consumer"
)

func Handler(message *sqs.Message) error {
	fmt.Println("handle", message)

	return nil
}

func main() {
	worker := consumer.New(&consumer.Config{
		Region:   aws.String("us-east-1"),
		QueueUrl: aws.String("https://sqs.us-east-1.amazonaws.com/130132914922/sqs-like_update_liked_videos-dev"),
	}, nil)

	worker.On(consumer.EventReceiveMessage, consumer.OnReceiveMessage(func(messages []*sqs.Message) {
		fmt.Println("OnReceiveMessage", messages)
	}))
	worker.On(consumer.EventProcessMessage, consumer.OnProcessMessage(func(message *sqs.Message) {
		fmt.Println("OnProcessMessage", message)
	}))
	worker.On(consumer.EventReceiveMessageError, consumer.OnReceiveMessageError(func(err error) {
		fmt.Println("OnReceiveMessageError", err)
	}))

	go worker.Start(Handler)

	worker.Concurrent(Handler, 6)

	select {}
}
