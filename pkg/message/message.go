package message

import "errors"

// Base interface for containers with messages.
type Serializer interface {
	ParseMessage(data []byte, message interface{}) error
	SerializeMessage(message interface{}) ([]byte, error)
	Format() string
	ContentMime() string
}

type SerializerBase struct{}

func (s *SerializerBase) ParseMessage(data []byte, message interface{}) error {
	return errors.New("unsupported method")
}

func (s *SerializerBase) SerializeMessage(message interface{}) ([]byte, error) {
	return nil, errors.New("unsupported method")
}

func (s *SerializerBase) Format() string {
	return ""
}

func (s *SerializerBase) ContentMime() string {
	return ""
}
