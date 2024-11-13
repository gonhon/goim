package kafka

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

func read(ctx context.Context) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaIP},
		Topic:          topic,
		CommitInterval: 1 * time.Second,
		GroupID:        "rec_team",
		StartOffset:    kafka.FirstOffset,
	})

	for {
		if message, err := reader.ReadMessage(ctx); err != nil {
			fmt.Println("读kafka失败: ", err)
			break
		} else {
			fmt.Printf("topic=%s, partition=%d, offset=%d, key=%s, value=%s \n", message.Topic, message.Partition, message.Offset, string(message.Key), string(message.Value))
		}
	}
}

func listenSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	fmt.Println("接收到信号: ", sig.String())
	if reader != nil {
		reader.Close()
	}
	os.Exit(0)
}
