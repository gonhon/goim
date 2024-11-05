package kafka

import (
	"github.com/segmentio/kafka-go"
)

var (
	reader *kafka.Reader
	topic  = "user_click"
)
