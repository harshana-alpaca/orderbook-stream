package binance

import (
	"MarketDataHub/internal/domain/orderbook/model"
)

type depthSnapshot struct {
	LastUpdateID uint64      `json:"lastUpdateID"`
	Bids         [][2]string `json:"bids"`
	Asks         [][2]string `json:"asks"`
	Symbol       string      `json:"symbol,omitempty"`
}

type update struct {
	EventType     string      `json:"e"`
	EventTime     uint64      `json:"E"`
	Symbol        string      `json:"s"`
	FirstUpdateID uint64      `json:"U"`
	LastUpdateID  uint64      `json:"u"`
	Bids          [][2]string `json:"b"`
	Asks          [][2]string `json:"a"`
}

func parseSnapshot(snapshot depthSnapshot) model.Snapshot {
	return model.Snapshot{
		Symbol:       snapshot.Symbol,
		Bids:         snapshot.Bids,
		Asks:         snapshot.Asks,
		LastUpdateId: snapshot.LastUpdateID,
	}
}

func parseUpdate(update update) model.Update {
	return model.Update{
		Symbol:        update.Symbol,
		Bids:          update.Bids,
		Asks:          update.Asks,
		LastUpdateId:  update.LastUpdateID,
		FirstUpdateId: update.FirstUpdateID,
	}
}
