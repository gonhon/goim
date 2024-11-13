package kafka

import (
	"github.com/segmentio/kafka-go"
)

var (
	reader  *kafka.Reader
	topic   = "user_click"
	kafkaIP = "10.66.0.11:9092"
)
