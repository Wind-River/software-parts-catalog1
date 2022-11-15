# communication-bus

## Writing Messages
```golang
package main

import (
	"encoding/json"
	"math"

	apachepulsar "github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.devstar.cloud/ip-systems/communication-bus.git/pulsar"
)

func main() {
	logger := log.With().Str(zerolog.CallerFieldName, "pulsar").Logger()

	client, err := pulsar.NewPulsarClient(&logger, "localhost", pulsar.DefaultPort, "listener-name", false)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	producer, err := pulsar.NewPulsarProducer(client, "topic-name")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer producer.Close() // Since preventAutoClosing was false, closing reader will close client

	for i := 0; i < math.MaxInt; i++ {
		payload, err := json.Marshal(i)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		messageID, err := producer.Send(payload)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		pulsarMessageID := messageID.(apachepulsar.MessageID)
		log.Trace().Interface("messageID", pulsarMessageID).Int("i", i).Msg("Sent Message")
	}
}
```
## Reading Messages
```golang
package main

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.devstar.cloud/ip-systems/communication-bus.git/pulsar"
)

func consume(payload []byte) error {
	var message string
	if err := json.Unmarshal(payload, &message); err != nil {
		return err
	}

	log.Info().Str(message, "message").Msg("Received Message")
	return nil
}

func main() {
	logger := log.With().Str(zerolog.CallerFieldName, "pulsar").Logger()

	client, err := pulsar.NewPulsarClient(&logger, "localhost", pulsar.DefaultPort, "listener-name", false)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	reader, err := pulsar.NewPulsarReader(client, "topic-name")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer reader.Close() // Since preventAutoClosing was false, closing reader will close client

	if err := reader.Read(context.Background(), consume, pulsar.Earliest); err != nil {
		log.Fatal().Err(err).Send()
	}
}
```