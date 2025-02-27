package pubsub_redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"github.com/redis/go-redis/v9"
)

const Provider string = "redis"

type RedisConfig struct {
	Host     string `default:"localhost" validate:"required" vmessage:"Host of Redis server not set"`
	Port     uint16 `default:"6379" validate:"gt=0" vmessage:"Port of Redis can not be zero"`
	Db       int
	Password string `mask:"true"`
}

type RedisClient struct {
	RedisConfig
	redisClient *redis.Client
	context     context.Context
}

func (r *RedisClient) Config() interface{} {
	return &r.RedisConfig
}

func (r *RedisClient) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	err := object_config.LoadLogValidate(cfg, log, vld, r, "pubsub", configPath...)
	if err != nil {
		return log.PushFatalStack("failed to init Redis client", err)
	}

	err = r.InitWithConfig(log, &r.RedisConfig)
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) InitWithConfig(log logger.Logger, cfg *RedisConfig) error {

	r.RedisConfig = *cfg

	address := fmt.Sprintf("%s:%d", r.Host, r.Port)
	r.context = context.Background()
	r.redisClient = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: r.Password,
		DB:       r.Db,
	})
	err := r.redisClient.Ping(r.context).Err()
	if err != nil {
		return log.PushFatalStack("failed to connect to Redis server", err)
	}
	return nil
}

func (p *RedisClient) Shutdown(ctx context.Context) error {
	return p.redisClient.Close()
}

//---------------------------------------

type Publisher struct {
	RedisClient
	pubsub.PublisherBase
}

func NewPublisher(serializer ...message.Serializer) *Publisher {
	p := &Publisher{}
	p.Construct(serializer...)
	return p
}

func (p *Publisher) Publish(topicName string, obj interface{}) error {

	payload, err := p.Serialize(obj)
	if err != nil {
		return err
	}

	return p.redisClient.Publish(p.context, topicName, payload).Err()
}

//---------------------------------------

type Subscriber struct {
	RedisClient
	pubsub_subscriber.SubscriberBase

	mutex    sync.RWMutex
	channels map[string]*redis.PubSub
}

func NewSubscriber(app app_context.Context, serializer ...message.Serializer) *Subscriber {
	s := &Subscriber{}
	s.Construct(app, serializer...)
	s.channels = make(map[string]*redis.PubSub)
	return s
}

func (s *Subscriber) Subscribe(topic pubsub_subscriber.Topic) (string, error) {

	subscriptionId, err := s.AddTopic(topic)
	if err != nil {
		return "", err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.channels[topic.Name()]
	if exists {
		return subscriptionId, nil
	}

	channel := s.redisClient.Subscribe(s.context, topic.Name())
	_, err = channel.Receive(s.context)
	if err != nil {
		return "", fmt.Errorf("failed to receive from redis pubsub channel: %s", err)
	}

	s.channels[topic.Name()] = channel

	ch := channel.Channel()
	readMessages := func() {
		for msg := range ch {
			opCtx := s.NewOpContext(topic.Name())
			s.Handle(opCtx, topic.Name(), []byte(msg.Payload))
			opCtx.Close()
		}
	}
	go readMessages()

	return subscriptionId, nil
}

func (s *Subscriber) Unsubscribe(topicName string, subscriptionId ...string) {

	unsubscribe := s.DeleteTopic(topicName, subscriptionId...)
	if !unsubscribe {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	channel, ok := s.channels[topicName]
	if !ok {
		return
	}

	channel.Unsubscribe(s.context)
	delete(s.channels, topicName)
}
