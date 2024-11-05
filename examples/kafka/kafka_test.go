package kafka

import (
	"context"
	"testing"
)

func TestKafkaConsumer(t *testing.T) {
	ctx := context.Background()
	go listenSignal()
	read(ctx)
}

func TestKafkaProducer(t *testing.T) {
	ctx := context.Background()
	Write(ctx)
}
func TestKafkaProducer1(t *testing.T) {
	Write1()
}
