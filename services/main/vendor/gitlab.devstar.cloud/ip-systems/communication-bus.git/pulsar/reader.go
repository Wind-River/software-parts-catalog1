package pulsar

import (
	"context"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
	bus "gitlab.devstar.cloud/ip-systems/communication-bus.git"
)

type PulsarReader struct {
	PulsarClient
	topic string
}

func NewPulsarReader(client *PulsarClient, topic string) (*PulsarReader, error) {
	return &PulsarReader{
		PulsarClient: *client,
		topic:        topic,
	}, nil
}

func (reader *PulsarReader) Read(ctx context.Context, consume bus.Consume, streamOffset bus.StreamOffset) error {
	var messageID pulsar.MessageID
	switch t := streamOffset.(type) {
	case pulsar.MessageID:
		messageID = t
	default:
		return fmt.Errorf("unknown streamOffset: %#v", t)
	}

	r, err := reader.CreateReader(pulsar.ReaderOptions{
		Topic:          reader.topic,
		StartMessageID: messageID,
	})
	if err != nil {
		return err
	}
	defer r.Close()

	for {
		message, err := r.Next(ctx)
		if err != nil {
			return err
		}

		if err := consume(message.Payload()); err != nil {
			return err
		}
	}
}

func (reader *PulsarReader) Close() error {
	reader.PulsarClient.Close()
	return nil
}
