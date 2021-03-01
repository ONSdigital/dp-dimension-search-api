package kafkaadapter

import kafka "github.com/ONSdigital/dp-kafka/v2"

// NewProducerAdapter returns a Producer object containing Kafka Producer
func NewProducerAdapter(producer *kafka.Producer) *Producer {
	return &Producer{kafkaProducer: producer}
}

// Producer exposes an output function, to satisfy the interface used by dp-kafka
type Producer struct {
	kafkaProducer *kafka.Producer
}

// Output mirrors the kafka producer output channel
func (p Producer) Output() chan []byte {
	return p.kafkaProducer.Channels().Output
}
