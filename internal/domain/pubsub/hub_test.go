package pubsub

import (
	"MarketDataHub/internal/domain/orderbook/model"
	"fmt"
	"testing"
	"time"
)

func TestPublishEndpoint(t *testing.T) {
	hub := NewHub()
	go nonBlockingListener("btcusdt", hub)
	go blockingListener("btcusdt", hub)
	time.Sleep(1 * time.Second)
	for i := 0; i < 50; i++ {
		update := model.Update{
			Symbol:        "btcusdt",
			FirstUpdateId: uint64(i),
			LastUpdateId:  uint64(i + 1),
		}
		i++
		err := hub.Publish("btcusdt", update)
		time.Sleep(100 * time.Millisecond)
		if err != nil {
			t.Error(err)
		}
	}
	time.Sleep(1 * time.Second)
}

func nonBlockingListener(symbol string, hub *Hub) {
	fmt.Println("starting listener")
	ch := make(chan model.Update)
	rst := make(chan bool)
	_ = hub.Subscribe(symbol, ch, rst)
	count := 0
	for update := range ch {
		fmt.Println("non blocking listener", update.FirstUpdateId, update.LastUpdateId)
		count++
	}
	fmt.Println("total events received", count)
}

func blockingListener(symbol string, hub *Hub) {
	fmt.Println("starting blocking listener")
	ch := make(chan model.Update, 100)
	rst := make(chan bool)
	_ = hub.Subscribe(symbol, ch, rst)
	count := 0
	for update := range ch {
		fmt.Println("blocking listener", update.FirstUpdateId, update.LastUpdateId)
		time.Sleep(100 * time.Millisecond)
		count++
	}
	fmt.Println("total events received", count)
}
