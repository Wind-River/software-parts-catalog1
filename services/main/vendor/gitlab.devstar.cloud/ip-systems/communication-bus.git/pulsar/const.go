package pulsar

import (
	apachepulsar "github.com/apache/pulsar-client-go/pulsar"
	bus "gitlab.devstar.cloud/ip-systems/communication-bus.git"
)

const DefaultPort int = 6650

var Earliest bus.StreamOffset = apachepulsar.EarliestMessageID()
var Latest bus.StreamOffset = apachepulsar.LatestMessageID()
