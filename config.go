package consumer

type Config struct {
	Region *string `json:"region"`

	AttributeNames          []*string `json:"attributeNames"`      // ["All"]
	MaxNumberOfMessages     *int64    `json:"maxNumberOfMessages"` // 1 - 10, 1
	MessageAttributeNames   []*string `json:"messageAttributeNames"`
	QueueUrl                *string   `json:"queueUrl"`
	ReceiveRequestAttemptId *string   `json:"receiveRequestAttemptId`
	VisibilityTimeout       *int64    `json:"visibilityTimeout"` // 0 - 43200, 30
	WaitTimeSeconds         *int64    `json:"waitTimeSeconds"`   // 0 - 20, 0, 10

	Idle  *int64 `json:"idle"`
	Sleep *int64 `json:"sleep"`
}
