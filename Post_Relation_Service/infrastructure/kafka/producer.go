package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase/interfacesUsecase"
	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer initializes a new Kafka writer
// brokers: e.g., []string{"localhost:9092"}
func NewKafkaProducer(brokers []string) interfacesUsecase.KafkaProducer {
	return &kafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
			// Create topic automatically if it doesn't exist
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *kafkaProducer) PublishEvent(topic string, message interface{}) error {
	// 1. Convert message to JSON bytes
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal kafka message: %w", err)
	}

	// 2. Write to Kafka
	err = p.writer.WriteMessages(context.Background(),
		kafka.Message{
			Topic: topic,
			Value: payload,
			Time:  time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}

	return nil
}
