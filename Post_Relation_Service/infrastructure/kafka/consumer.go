package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/events"
	interfacesUsecase "github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase/interfacesUsecase"
	"github.com/segmentio/kafka-go"
)

func StartFeedConsumer(
	brokerStr string,
	topic string,
	groupID string,
	feedUsecase interfacesUsecase.FeedUsecase,
	caCert, accessCert, accessKey []byte,
) {

	cert, _ := tls.X509KeyPair(accessCert, accessKey)

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
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(brokerStr, ","),
		Topic:    topic,
		GroupID:  groupID,
		Dialer:   dialer,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Println("Kafka read error:", err)
			continue
		}

		var base struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal(m.Value, &base); err != nil {
			continue
		}

		switch base.Type {

		case "POST_CREATED":
			var event events.PostCreatedEvent

			if err := json.Unmarshal(m.Value, &event); err != nil {
				fmt.Println("Failed to parse POST_CREATED:", err)
				continue
			}

			err := feedUsecase.ProcessPostCreated(event)
			if err != nil {
				fmt.Println("Feed processing failed:", err)
			}
		}
	}
}
