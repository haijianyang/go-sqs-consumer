# go-sqs-consumer
SQS Consumer &amp; Worker for Go

## Install

```console
go get github.com/haijianyang/go-sqs-consumer
```

## Quick Start

```go
// create worker for the queue
worker := consumer.New(&consumer.Config{
	Region:   aws.String("region"),
	QueueUrl: aws.String("url"),
}, nil)

// start worker for polling messages
worker.Start(func(message *sqs.Message) error {
	fmt.Println("handle", message)

	return nil
}
```

## Document

### Credentials

By default the consumer will look for AWS credentials in the places [specified by the AWS SDK](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html).

```go
// consumer will create SQS client for worker
worker := consumer.New(&consumer.Config{
	Region:   aws.String("region"),
	QueueUrl: aws.String("url"),
}, nil)
```

You can set a instance of the AWS SQS client when create worker.

```go
// init AWS session
sess, err := session.NewSession(&aws.Config{
	Region: aws.String("region")},
)
// init SQS client
svr = sqs.New(sess)

// set SQS client
worker := consumer.New(&consumer.Config{
	Region:   aws.String("region"),
	QueueUrl: aws.String("url"),
}, svr)
```

### Events

```go
worker.On(consumer.EventReceiveMessage, consumer.OnReceiveMessage(func(messages []*sqs.Message) {
	fmt.Println("OnReceiveMessage", messages)
}))

worker.On(consumer.EventProcessMessage, consumer.OnProcessMessage(func(message *sqs.Message) {
		fmt.Println("OnProcessMessage", message)
}))

worker.On(consumer.EventReceiveMessageError, consumer.OnReceiveMessageError(func(err error) {
	fmt.Println("OnReceiveMessageError", err)
}))
```

### Concurrent

```go
func Handler(message *sqs.Message) error {
	fmt.Println("handle", message)

	return nil
}

worker := consumer.New(&consumer.Config{
	Region:   aws.String("region"),
	QueueUrl: aws.String("url"),
}, nil)

// 1. do it yourself
go worker.Start(Handler)
go worker.Start(Handler)
...

// 2. automation
concurrency := 6
worker.Concurrent(Handler, concurrency)
```

## API

### consumer.Config
* `Region` - _String_ - The AWS region.
* `AttributeNames` - _[]String_ - List of queue attributes to retrieve (i.e. ['All', 'ApproximateFirstReceiveTimestamp', 'ApproximateReceiveCount']).
* `MaxNumberOfMessages` - _Int64_ - The maximum number of messages to return. Valid values: 1 to 10. Default: 1.
* `MessageAttributeNames` - _[]String_ - The name of the message attribute.
* `QueueUrl` - _String_ - The URL of the Amazon SQS queue from which messages are received.
* `ReceiveRequestAttemptId` - _String_ - This parameter applies only to FIFO (first-in-first-out) queues.
* `VisibilityTimeout` - _Int64_ - The duration (in seconds) that the received messages are hidden from subsequent retrieve requests after being retrieved by a ReceiveMessage request. Default: 30.
* `WaitTimeSeconds` - _Int64_ - The duration (in seconds) for which the call waits for a message to arrive in the queue before returning.
* `Idle` - _Int_ - The number of worker to sleep.
* `Sleep` - _Int_ - The duration (in seconds) of worker sleep.

### consumer.New(config *Config, svr *sqs.SQS)

Creates a new SQS worker.

* `config` - _Config_ - The consumer.Config, configuration for worker
* `svr` - _*sqs.SQS_ - The AWS SQS client

### worker.Start(handler Handler)

Start polling the queue for messages.

* `handler` - _func(message *sqs.Message) error_ - To be called whenever a message is received

### worker.On(event string, callback interface{})

Register event listener.

* `event` - _String_ - The worker events.
* `callback` - _Func_ - The worker event callback functions.

#### Worker Events

|Event|Func|Description|
|-----|------|-----------|
|`EventReceiveMessage`|OnReceiveMessage func(messages []*sqs.Message)|Fired when messages is received from SQS queue.|
|`EventProcessMessage`|OnProcessMessage func(message *sqs.Message)|Fired when a message is successfully processed and removed from the queue.|
|`OnReceiveMessageError`|OnReceiveMessageError func(err error)|Fired when receiving messages from SQS queue fail.|
