package pubsub

import (
	"MarketDataHub/internal/domain/orderbook/model"
	"log"
	"sync"
)

// Hub implements the pub-sub mechanism in the market data hub.
// It is expected that subscribers needs to manage a non-blocking (buffered) channel.
// For Publisher, hub will maintain a channel per topic.
type Hub struct {
	// map of map to store subscribers
	subscribers map[string]map[chan<- model.Update]subscription
	lock        sync.RWMutex
}

type subscription struct {
	events chan<- model.Update // shares data
	rst    chan<- bool         // sends control signal.
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string]map[chan<- model.Update]subscription),
	}
}

func (h *Hub) Subscribe(topic string, ch chan<- model.Update, control chan<- bool) error {
	log.Println(topic, "subscriber detected, waiting for lock acquire")
	h.lock.Lock()
	log.Println(topic, "subscriber lock acquired")
	defer h.lock.Unlock()
	if _, ok := h.subscribers[topic]; !ok {
		h.subscribers[topic] = make(map[chan<- model.Update]subscription)
	}
	h.subscribers[topic][ch] = subscription{ch, control}
	return nil
}

func (h *Hub) Unsubscribe(topic string, ch chan<- model.Update) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	if _, ok := h.subscribers[topic]; !ok {
		return nil
	}
	delete(h.subscribers[topic], ch)
	return nil
}

func (h *Hub) Publish(topic string, event model.Update) error {
	log.Println(topic, "publishing event")
	h.lock.RLock()
	subscribers, ok := h.subscribers[topic]
	if !ok {
		log.Println(topic, "publishing event : early return")
		h.lock.RUnlock()
		return nil
	}
	subs := make([]subscription, 0, len(subscribers))
	for _, v := range subscribers {
		subs = append(subs, v)
	}
	h.lock.RUnlock()
	slowSubs := make([]subscription, 0, len(subs))

	for _, s := range subs {
		log.Println(topic, "publishing event to sub", event.FirstUpdateId)
		select {
		case s.events <- event:
		default:
			select {
			case s.rst <- true:
			default:
			}
			// indicate that it's dropped, and delete the subscription. Don't keep slow channels.
			slowSubs = append(slowSubs, s)
		}
	}
	h.lock.Lock()
	for _, sub := range slowSubs {
		delete(h.subscribers[topic], sub.events)
	}
	if len(h.subscribers[topic]) == 0 {
		delete(h.subscribers, topic)
	}
	h.lock.Unlock()
	return nil
}
