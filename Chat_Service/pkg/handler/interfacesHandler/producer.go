package interfacesHandler

// KafkaProducer defines the contract for sending events to the message broker.
// This allows the usecase to remain agnostic of the specific library (Sarama, Segmentio, etc.)
type KafkaProducer interface {
	PublishEvent(topic string, message interface{}) error
}
