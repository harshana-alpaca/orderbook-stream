package exchange

import (
	"MarketDataHub/internal/domain/orderbook/model"
	"context"
)

type Streamer interface {
	Stream(ctx context.Context, symbol string) (model.Snapshot, <-chan model.Update, error)
}
