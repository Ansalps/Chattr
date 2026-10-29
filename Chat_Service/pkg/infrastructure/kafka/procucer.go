package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Ansalps/Chattr_Chat_Service/pkg/handler/interfacesHandler"
	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
	dialer  *kafka.Dialer // ✅ add this
	brokers []string      // ✅ add this
}

// NewKafkaProducer initializes a new Kafka writer
// brokers: e.g., []string{"localhost:9092"}
func NewKafkaProducer(brokers []string, caCert, accessCert, accessKey []byte) interfacesHandler.KafkaProducer {
	// 1. Setup TLS
	cert, err := tls.X509KeyPair(accessCert, accessKey)
	if err != nil {
		log.Fatalf("KAFKA: Failed to load client cert/key: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	// 2. Shared Transport & Dialer
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		TLS:       tlsConfig,
		ClientID:  "chattr-producer",
	}

	// 3. Quick Connection Test (Bootstrapping)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	metaConn, err := dialer.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		log.Fatalf("KAFKA: Failed to connect to broker: %v", err)
	}
	defer metaConn.Close()
	log.Println("KAFKA AUTH SUCCESS: TLS Handshake complete.")

	partitions, err := metaConn.ReadPartitions()
	if err != nil {
		log.Fatalf("KAFKA: Failed to read cluster partitions: %v", err)
	}

	// Collect existing topics.
	existingTopics := make(map[string]bool)
	for _, partition := range partitions {
		existingTopics[partition.Topic] = true
	}

	// Create chat-events if it doesn't exist.
	topic := "chat-events"

	if existingTopics[topic] {
		log.Printf("KAFKA: Topic '%s' already exists", topic)
	} else {
		log.Printf("KAFKA: Topic '%s' not found. Attempting to create it...", topic)

		err = metaConn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
		if err != nil {
			log.Fatalf("KAFKA: Failed to create topic '%s': %v", topic, err)
		}

		log.Printf("KAFKA: Successfully created topic '%s'", topic)
	}

	conn, err := dialer.DialLeader(
		ctx,
		"tcp",
		brokers[0],
		topic,
		0, // partition
	)
	if err != nil {
		log.Fatalf("KAFKA: Topic '%s' is not reachable: %v", topic, err)
	}
	conn.Close()

	log.Printf("KAFKA: Topic '%s' is reachable", topic)
	// 4. Return Producer with a dynamic Writer
	return &kafkaProducer{
		dialer:  dialer,
		brokers: brokers,
	}
}

func (p *kafkaProducer) PublishEvent(topic string, message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal kafka message: %w", err)
	}

	conn, err := p.dialer.DialLeader(
		context.Background(),
		"tcp",
		p.brokers[0],
		topic,
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %w", err)
	}
	defer conn.Close()

	_, err = conn.WriteMessages(
		kafka.Message{
			Value: payload,
			Time:  time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}

	log.Printf("KAFKA: Event published successfully to topic: %s", topic)
	return nil
}
