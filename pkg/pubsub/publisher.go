package pubsub

import (
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/message/message_json"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Publisher interface {
	Publish(topicName string, obj interface{}) error
	Shutdown()
}

type PublisherBase struct {
	serializer message.Serializer
}

func (p *PublisherBase) Construct(serializer ...message.Serializer) {
	p.serializer = utils.OptionalArg(message.Serializer(message_json.Serializer), serializer...)
}

func (p *PublisherBase) Serialize(msg interface{}) ([]byte, error) {
	return p.serializer.SerializeMessage(msg)
}
