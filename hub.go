/*
 * @file: hub.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package mhub

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/jquiterio/uuid"
)

var ctx = context.Background()

type Hub struct {
	Subscribers []Subscriber
	Topics      []string
	Registry    *redis.Client
}

type Subscriber struct {
	ID     string
	Topics []string
}

func (m *Message) ToMap() map[string]any {
	return map[string]any{
		"subscriber_id": m.SubscriberID,
		"id":            m.ID,
		"topic":         m.Topic,
		"data":          m.Payload,
	}
}

func NewHub() *Hub {
	conf := GetDefaultConfig()
	return &Hub{
		Subscribers: make([]Subscriber, 0),
		Topics:      make([]string, 0),
		Registry: redis.NewClient(&redis.Options{
			Addr: conf.Redis.Addr,
			DB:   conf.Redis.DB,
		}),
	}
}

func NewHubWithConfig(conf Config) *Hub {
	return &Hub{
		Subscribers: make([]Subscriber, 0),
		Topics:      make([]string, 0),
		Registry: redis.NewClient(&redis.Options{
			Addr: conf.Redis.Addr,
			DB:   conf.Redis.DB,
		}),
	}
}

func (h *Hub) Subscribe(sub Subscriber) {
	h.Subscribers = append(h.Subscribers, sub)
	h.addTopicFromSubscribers()
}

func (h *Hub) removeTopicFromSubscribers() {
	for _, sub := range h.Subscribers {
		for i, t := range sub.Topics {
			if !h.HasTopic(t) {
				sub.Topics = append(sub.Topics[:i], sub.Topics[i+1:]...)
			}
		}
	}
}

func (h *Hub) HasTopic(topic string) bool {
	for _, t := range h.Topics {
		if t == topic {
			return true
		}
	}
	return false
}

func (h *Hub) Unsubscribe(sub *Subscriber, topics []string) (ok bool) {
	for _, topic := range topics {
		sub.RemoveTopic(topic)
	}
	h.removeTopicFromSubscribers()
	return true
}

func (h *Hub) GetSubscriber(id string) *Subscriber {
	for _, sub := range h.Subscribers {
		if sub.ID == id {
			return &sub
		}
	}
	return nil
}

func (h *Hub) addTopicFromSubscribers() {
	for _, sub := range h.Subscribers {
		h.Topics = append(h.Topics, sub.Topics...)
	}
}

func NewSubscriber(topics ...string) *Subscriber {
	return &Subscriber{
		ID:     uuid.New().String(),
		Topics: topics,
	}
}

func (s Subscriber) HasTopic(topic string) bool {
	for _, t := range s.Topics {
		if t == topic {
			return true
		}
	}
	return false
}

func (s *Subscriber) AddTopic(topic string) {
	s.Topics = append(s.Topics, topic)
}

func (s *Subscriber) RemoveTopic(topic string) {
	for i, t := range s.Topics {
		if t == topic {
			s.Topics = append(s.Topics[:i], s.Topics[i+1:]...)
			return
		}
	}
}

func (h *Hub) Publish(msg Message) error {
	return h.Registry.Publish(ctx, msg.Topic, msg.Payload).Err()
}
