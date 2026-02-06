package main

import (
	"MarketDataHub/internal/client"
	"MarketDataHub/internal/domain/orderbook"
	"MarketDataHub/internal/domain/pubsub"
	"MarketDataHub/internal/exchange"
	"MarketDataHub/internal/exchange/binance"
	"MarketDataHub/internal/store/memory"
	"context"
)

func main() {
	var connector exchange.Streamer = binance.NewStreamer()
	var memoryStore orderbook.Keeper = memory.NewStore()
	hub := pubsub.NewHub()
	var service = orderbook.NewService(connector, memoryStore, hub, []string{"btcusdt", "bnbusdt"})
	_ = service.Start(context.Background())
	client.StartServer(hub, *service)
}
