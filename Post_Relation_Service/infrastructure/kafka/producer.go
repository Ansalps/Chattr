package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase/interfacesUsecase"
	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
	dialer  *kafka.Dialer   // ✅ add this
	brokers []string        // ✅ add this
}

func NewKafkaProducer(brokers []string, caCert, accessCert, accessKey []byte) interfacesUsecase.KafkaProducer {
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
	// Using one transport prevents the "Unknown Topic" error caused by metadata lag
	sharedTransport := &kafka.Transport{
		TLS: tlsConfig,
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		},
	}

	// 3. Quick Connection Test (Bootstrapping)
	conn, err := sharedTransport.Dial(context.Background(), "tcp", brokers[0])
	if err != nil {
		log.Fatalf("KAFKA AUTH ERROR: Connection failed: %v", err)
	}
	conn.Close()
	log.Println("KAFKA AUTH SUCCESS: TLS Handshake complete.")

	conn, err = dialer.DialLeader(
		context.Background(),
		"tcp",
		brokers[0],
		"post-events",
		0, // partition
	)
	if err != nil {
		log.Fatalf("❌ Topic NOT reachable: %v", err)
	}
	defer conn.Close()
	
	log.Println("✅ Topic is reachable and partition exists")
	// 4. Return Producer with a dynamic Writer
	return &kafkaProducer{
		dialer:dialer,
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
