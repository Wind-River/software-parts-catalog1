package pulsar

import (
	"context"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/pkg/errors"
)

type PulsarProducer struct {
	client   *PulsarClient
	topic    string
	producer pulsar.Producer
}

func NewPulsarProducer(client *PulsarClient, topic string) (*PulsarProducer, error) {
	ret := new(PulsarProducer)

	ret.topic = topic
	ret.client = client

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "could not create producer")
	}
	ret.producer = producer

	return ret, nil
}

func (producer PulsarProducer) Send(payload []byte) (messageID interface{}, err error) {
	return producer.producer.Send(context.Background(), &pulsar.ProducerMessage{
		Payload: payload,
	})
}

func (producer *PulsarProducer) Close() error {
	producer.producer.Close()
	producer.client.Close()

	return nil
}
