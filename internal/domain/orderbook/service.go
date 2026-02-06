package orderbook

import (
	"MarketDataHub/internal/domain/orderbook/model"
	"MarketDataHub/internal/domain/pubsub"
	"MarketDataHub/internal/exchange"
	"context"
	"fmt"
)

type Service struct {
	feed    exchange.Streamer
	keeper  Keeper
	hub     pubsub.Publisher
	symbols []string
}

func NewService(conn exchange.Streamer, keeper Keeper, hub pubsub.Publisher, symbols []string) *Service {
	return &Service{
		feed:    conn,
		keeper:  keeper,
		hub:     hub,
		symbols: symbols,
	}
}

func (s *Service) Start(ctx context.Context) error {
	for _, symbol := range s.symbols {
		go s.startListeningToSymbol(ctx, symbol)
	}
	return nil
}

func (s *Service) GetSnapshot(ctx context.Context, symbol string) (model.Snapshot, error) {
	return s.keeper.GetSnapshot(ctx, symbol)
}
func (s *Service) startListeningToSymbol(ctx context.Context, symbol string) {
	// we are now in a separate go routine. run a loop to listen to updates.
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		//start listening.
		snapshot, updates, err := s.feed.Stream(ctx, symbol)
		if err != nil {
			//restart the connector.
			continue
		}
		changes, err := s.keeper.ApplySnapshot(ctx, symbol, snapshot)
		if err != nil {
			continue
		}
		fmt.Println("writing to hub")
		err = s.hub.Publish(symbol, changes)
		if err != nil {
			continue
		}
		fmt.Println("listening on updates to hub")
		restart := s.publishAndStore(ctx, symbol, updates)
		if !restart {
			return
		}
	}

}

func (s *Service) publishAndStore(ctx context.Context, symbol string, updates <-chan model.Update) bool {
	for {
		select {
		case <-ctx.Done():
			// we don't want to restart this.
			return false
		case update, ok := <-updates:
			if !ok {
				// Channel is closed from the connector. Need to handle it.
				return true
			}
			err := s.hub.Publish(symbol, update)
			if err != nil {
				return true
			}
			err = s.keeper.ApplyUpdate(ctx, symbol, update)
			// Cannot apply the update.
			if err != nil {
				return true
			}
		}
	}
}
