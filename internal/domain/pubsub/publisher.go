package pubsub

import (
	"MarketDataHub/internal/domain/orderbook/model"
)

type Publisher interface {
	Publish(topic string, event model.Update) error
}
