package consumer

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Handler func(context context.Context, message *sqs.Message) error

type Worker struct {
	Config *Config
	Input  *sqs.ReceiveMessageInput
	Sqs    *sqs.SQS
	Events map[string]interface{}
}

func New(config *Config, sqs *sqs.SQS) *Worker {
	worker := &Worker{
		Config: config,
		Sqs:    sqs,
		Events: make(map[string]interface{}),
	}

	worker.SetConfig(*config)

	if worker.Sqs == nil {
		worker.Sqs = NewClient(worker.Config.Region)
	}

	return worker
}

func (this *Worker) SetConfig(config Config) {
	this.Config = &config

	if config.Idle == nil {
		config.Idle = aws.Int64(0)
	}
	if config.Sleep == nil {
		config.Sleep = aws.Int64(0)
	}

	if config.AttributeNames == nil {
		config.AttributeNames = []*string{aws.String("All")}
	}
	if config.MaxNumberOfMessages == nil {
		config.MaxNumberOfMessages = aws.Int64(1)
	}
	if config.VisibilityTimeout == nil {
		config.VisibilityTimeout = aws.Int64(30)
	}
	if config.WaitTimeSeconds == nil {
		config.WaitTimeSeconds = aws.Int64(10)
	}

	this.Input = &sqs.ReceiveMessageInput{
		AttributeNames:          config.AttributeNames,
		MaxNumberOfMessages:     config.MaxNumberOfMessages,
		MessageAttributeNames:   config.MessageAttributeNames,
		QueueUrl:                config.QueueUrl,
		ReceiveRequestAttemptId: config.ReceiveRequestAttemptId,
		VisibilityTimeout:       config.VisibilityTimeout,
		WaitTimeSeconds:         config.WaitTimeSeconds,
	}
}

func (this *Worker) SetAttributeNames(attributeNames []string) {
	this.Input.AttributeNames = make([]*string, len(attributeNames))
	for i := 0; i < len(attributeNames); i++ {
		this.Input.AttributeNames[i] = aws.String(attributeNames[i])
	}
}

func (this *Worker) SetMaxNumberOfMessages(maxNumberOfMessages int64) {
	this.Input.MaxNumberOfMessages = aws.Int64(maxNumberOfMessages)
}

func (this *Worker) SetMessageAttributeNames(messageAttributeNames []string) {
	if len(messageAttributeNames) == 0 {
		this.Input.MessageAttributeNames = nil
	} else {
		this.Input.MessageAttributeNames = make([]*string, len(messageAttributeNames))
		for i := 0; i < len(messageAttributeNames); i++ {
			this.Input.MessageAttributeNames[i] = aws.String(messageAttributeNames[i])
		}
	}
}

func (this *Worker) SetQueueUrl(queueUrl string) {
	if len(queueUrl) == 0 {
		this.Input.QueueUrl = nil
	} else {
		this.Input.QueueUrl = aws.String(queueUrl)
	}
}

func (this *Worker) SetReceiveRequestAttemptId(receiveRequestAttemptId string) {
	if len(receiveRequestAttemptId) == 0 {
		this.Input.ReceiveRequestAttemptId = nil
	} else {
		this.Input.ReceiveRequestAttemptId = aws.String(receiveRequestAttemptId)
	}
}

func (this *Worker) SetVisibilityTimeout(visibilityTimeout int64) {
	this.Input.VisibilityTimeout = aws.Int64(visibilityTimeout)
}

func (this *Worker) SetWaitTimeSeconds(waitTimeSeconds int64) {
	this.Input.WaitTimeSeconds = aws.Int64(waitTimeSeconds)
}

func (this *Worker) SetSqs(sqs *sqs.SQS) {
	this.Sqs = sqs
}

func (this *Worker) On(event string, callback interface{}) {
	switch event {
	case EventReceiveMessage:
		cb, _ := callback.(OnReceiveMessage)
		if cb == nil {
			panic(errors.New("Error OnReceiveMessage"))
		}
		break
	case EventProcessMessage:
		cb, _ := callback.(OnProcessMessage)
		if cb == nil {
			panic(errors.New("Error OnProcessMessage"))
		}
		break
	case EventReceiveMessageError:
		cb, _ := callback.(OnReceiveMessageError)
		if cb == nil {
			panic(errors.New("Error OnReceiveMessageError"))
		}
		break
	default:
		panic(errors.New("Error Event"))
	}

	this.Events[event] = callback
}

func (this *Worker) Start(context context.Context, handler Handler) {
	for idle := int64(0); ; {
		if aws.Int64Value(this.Config.Idle) > 0 && idle > aws.Int64Value(this.Config.Idle) && aws.Int64Value(this.Config.Sleep) > 0 {
			idle = 0

			time.Sleep(time.Duration(aws.Int64Value(this.Config.Sleep)) * time.Second)
		}

		output, err := this.Sqs.ReceiveMessage(this.Input)
		if err != nil {
			if err != nil {
				log.Printf("[SQS] ReceiveMessage error: %v %v", err, output)
			}

			if this.Events[EventReceiveMessageError] != nil {
				this.Events[EventReceiveMessageError].(OnReceiveMessageError)(err)
			}

			continue
		}

		if len(output.Messages) > 0 {
			idle = 0

			if this.Events[EventReceiveMessage] != nil {
				this.Events[EventReceiveMessage].(OnReceiveMessage)(output.Messages)
			}

			this.run(context, handler, output.Messages)
		} else {
			idle++
		}
	}
}

func (this *Worker) Concurrent(context context.Context, handler Handler, concurrency int) {
	for i := 0; i < concurrency; i++ {
		go this.Start(context, handler)
	}
}

func (this *Worker) run(context context.Context, handler Handler, messages []*sqs.Message) {
	var wg sync.WaitGroup
	wg.Add(len(messages))

	for i := range messages {
		go func(message *sqs.Message) {
			defer wg.Done()

			if err := this.handleMessage(context, message, handler); err != nil {
				log.Printf("[SQS] handleMessage error: %v", err)
			} else {
				if this.Events[EventProcessMessage] != nil {
					this.Events[EventProcessMessage].(OnProcessMessage)(message)
				}
			}
		}(messages[i])
	}

	wg.Wait()
}

func (this *Worker) handleMessage(context context.Context, message *sqs.Message, handler Handler) error {
	if err := handler(context, message); err != nil {
		return err
	}

	input := &sqs.DeleteMessageInput{
		QueueUrl:      this.Input.QueueUrl,
		ReceiptHandle: message.ReceiptHandle,
	}

	output, err := this.Sqs.DeleteMessage(input)
	if err != nil {
		log.Printf("[SQS] DeleteMessage error: %v %v", err, output)
	}

	return nil
}
