// communication_bus defines the interfaces any stream implementation will conform to.
package communication_bus

import "context"

// A Producer can write messages to a stream.
type Producer interface {
	Send(payload []byte) (messageID interface{}, err error)
	Close() error
}

// Consume is a function that takes in and processes the raw bytes of a message.
type Consume func(payload []byte) error

// A Consumer can read a single message from the stream.
type Consumer interface {
	Consume
	Close() error
}

// StreamOffset is whatever the implementation needs to set a Reader offset.
// Typically includes at least eariest message and latest message.
type StreamOffset interface{}

// A Reader can read messages from a stream using the given Consume function.
// The first message is set using StreamOffset.
// Reader will stop when the given context is done.
type Reader interface {
	Read(context.Context, Consume, StreamOffset) error
	Close() error
}
