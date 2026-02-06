package orderbook

import (
	"MarketDataHub/internal/domain/orderbook/model"
	"context"
)

type Keeper interface {
	ApplySnapshot(ctx context.Context, symbol string, snapshot model.Snapshot) (model.Update, error)
	ApplyUpdate(ctx context.Context, symbol string, update model.Update) error
	GetSnapshot(ctx context.Context, symbol string) (model.Snapshot, error)
	DeleteSnapshot(ctx context.Context, symbol string) error
}
