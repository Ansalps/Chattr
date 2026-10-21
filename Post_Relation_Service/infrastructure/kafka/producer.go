package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase/interfacesUsecase"
	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
    writer  *kafka.Writer
    brokers []string
}

func NewKafkaProducer(brokers []string, caCert, accessCert, accessKey []byte) interfacesUsecase.KafkaProducer {
    // 1. Setup TLS (same as before)
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

    dialer := &kafka.Dialer{
        Timeout:   10 * time.Second,
        DualStack: true,
        TLS:       tlsConfig,
        ClientID:  "chattr-producer",
    }

	// 2. Quick Connection Test & Auto-Topic Creation (same as before)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    metaConn, err := dialer.DialContext(ctx, "tcp", brokers[0])
    if err != nil {
        log.Fatalf("❌ Failed to connect to broker for metadata: %v", err)
    }
    defer metaConn.Close()

	partitions, err := metaConn.ReadPartitions()
    if err != nil {
        log.Fatalf("❌ Failed to read cluster partitions: %v", err)
    }

    topicExists := false
    for _, p := range partitions {
        if p.Topic == "post-events" {
            topicExists = true
            break
        }
    }

	if !topicExists {
        log.Println("⚠️ Topic 'post-events' not found. Attempting to create it...")
        err = metaConn.CreateTopics(kafka.TopicConfig{
            Topic:             "post-events",
            NumPartitions:     1,
            ReplicationFactor: 1,
        })
        if err != nil {
            log.Fatalf("❌ Failed to create topic 'post-events': %v", err)
        }
        log.Println("✅ Successfully created topic 'post-events'")
    } else {
        log.Println("✅ Topic 'post-events' exists and is reachable")
    }

	// 3. Initialize a persistent kafka.Writer (Best Practice)
    writer := &kafka.Writer{
        Addr:         kafka.TCP(brokers...),
        Balancer:     &kafka.LeastBytes{}, // Distributes load efficiently
        Transport:    &kafka.Transport{TLS: tlsConfig},
        BatchTimeout: 10 * time.Millisecond, // Flushes batches quickly for low latency
        RequiredAcks: kafka.RequireOne,      // Appropriate for single-node / free tier
    }

    log.Println("✅ Kafka Producer initialized successfully with a shared Writer")

    return &kafkaProducer{
        writer:  writer,
        brokers: brokers,
    }
}

func (p *kafkaProducer) PublishEvent(topic string, message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal kafka message: %w", err)
	}

	// Create a context with a timeout for sending the message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: payload,
		Time:  time.Now(),
	})

	if err != nil {
		return fmt.Errorf("failed to write message to kafka topic %s: %w", topic, err)
	}

	log.Printf("KAFKA: Event published successfully to topic: %s", topic)
	return nil
}
