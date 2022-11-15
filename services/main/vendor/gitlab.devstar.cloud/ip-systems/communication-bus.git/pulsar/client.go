package pulsar

import (
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type PulsarClient struct {
	pulsar.Client
	logger     *zerolog.Logger
	references uint
}

// NewPulsarClient creates the client required for PulsarProducers or PulsarReaders.
// preventAutoClosing will prevent closing all of these Producers or Readers from also closing this client.
func NewPulsarClient(logger *zerolog.Logger, host string, port int, listener string, preventAutoClosing bool) (*PulsarClient, error) {
	if port <= 0 {
		port = DefaultPort
	}

	options := pulsar.ClientOptions{
		URL:               fmt.Sprintf("pulsar://%s:%d", host, port),
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		// Logger:            logger,
	}
	if listener == "" {
		options.ListenerName = listener
	}

	client, err := pulsar.NewClient(options)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create Pulsar client")
	}

	ret := new(PulsarClient)
	ret.logger = logger
	ret.Client = client
	if preventAutoClosing {
		// Prevents closing readers/producers from closing this client
		ret.references = 1
	}

	return ret, nil
}

func (client *PulsarClient) Logger() *zerolog.Logger {
	return client.logger
}

func (client *PulsarClient) Close() error {
	if client.references > 0 {
		client.references--
	}

	if client.references == 0 {
		client.Client.Close()
	}

	return nil
}

func (client *PulsarClient) CreateProducer(options pulsar.ProducerOptions) (pulsar.Producer, error) {
	ret, err := client.Client.CreateProducer(options)
	if err != nil {
		return ret, err
	}

	client.references++
	return ret, nil
}

func (client *PulsarClient) CreateReader(options pulsar.ReaderOptions) (pulsar.Reader, error) {
	ret, err := client.Client.CreateReader(options)
	if err != nil {
		return ret, err
	}

	client.references++
	return ret, nil
}

func (client *PulsarClient) Subscribe(options pulsar.ConsumerOptions) (pulsar.Consumer, error) {
	ret, err := client.Client.Subscribe(options)
	if err != nil {
		return ret, err
	}

	client.references++
	return ret, nil
}
