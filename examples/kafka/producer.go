package kafka

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

func Write(ctx context.Context) {
	writer := &kafka.Writer{
		// Addr:                   kafka.TCP("10.66.0.11:9092"),
		Addr:                   kafka.TCP("localhost:9092"),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		WriteTimeout:           1 * time.Second,
		RequiredAcks:           kafka.RequireNone,
		AllowAutoTopicCreation: true,
	}
	defer writer.Close()

	tm := time.Now().Unix()
	for i := 0; i < 3; i++ {
		if err := writer.WriteMessages(
			ctx,
			kafka.Message{Key: []byte("1"), Value: []byte("小" + strconv.Itoa(int(tm)))},
			kafka.Message{Key: []byte("2"), Value: []byte("白" + strconv.Itoa(int(tm)))},
			kafka.Message{Key: []byte("3"), Value: []byte("小" + strconv.Itoa(int(tm)))},
			kafka.Message{Key: []byte("1"), Value: []byte("熊" + strconv.Itoa(int(tm)))},
			kafka.Message{Key: []byte("1"), Value: []byte("猫" + strconv.Itoa(int(tm)))},
		); err != nil {
			if err == kafka.LeaderNotAvailable {
				time.Sleep(500 * time.Millisecond)
				continue
			} else {
				fmt.Printf("批量写kafka失败：%v \n", err)
			}
		} else {
			break
		}
	}
}
func Write1() {

	// 替换为你的 Kafka IP 地址
	kafkaIP := "127.0.0.1:9092" // 例如 "192.168.1.100:9092"

	// 创建 Kafka 生产者
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaIP},
		Topic:   "user_click", // 替换为你的目标主题
	})
	defer writer.Close()

	// 发送消息
	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("Key-A"),
			Value: []byte("Hello Kafka!"),
		},
	)
	if err != nil {
		log.Fatalf("Failed to write message: %v", err)
	}

	log.Println("Message sent successfully")
}
