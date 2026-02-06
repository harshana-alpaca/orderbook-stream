package pubsub

import "MarketDataHub/internal/domain/orderbook/model"

type Subscriber interface {
	Subscribe(topic string, updates chan<- model.Update, rst chan<- bool) error
	Unsubscribe(topic string, updates chan<- model.Update) error
}
