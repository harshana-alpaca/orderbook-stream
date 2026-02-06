package binance

import (
	"MarketDataHub/internal/domain/orderbook/model"
	"context"
	"fmt"
)

type Streamer struct{}

func NewStreamer() *Streamer {
	return &Streamer{}
}

func (c *Streamer) Stream(ctx context.Context, symbol string) (model.Snapshot, <-chan model.Update, error) {
	stop := make(chan bool)
	updates, err := getUpdates(ctx, symbol, stop)
	//updates are queued.
	if err != nil {
		fmt.Println(err)
		return model.Snapshot{}, nil, err
	}
	snapshot, err := getSnapshot(symbol)
	if err != nil {
		fmt.Println(err)
		return model.Snapshot{}, nil, err
	}
	lastUpdateSnapshot := snapshot.LastUpdateID

	returnCh := make(chan model.Update, 100)
	go writeValidUpdates(lastUpdateSnapshot, updates, stop, returnCh)
	return parseSnapshot(snapshot), returnCh, nil
}

func writeValidUpdates(lastUpdateId uint64, inbound <-chan update, stopInbound chan<- bool, outbound chan<- model.Update) {
	var found = false
	//what if first message actually after the rest call? lastUpdateID < u.FirstUpdateID
	//fmt.Println("writing updates")
	i := 0
	for u := range inbound {
		i++
		// discard if update is older than the snapshot.
		switch {
		case found:
			//fmt.Println("writing updates without processing last update id")
			outbound <- parseUpdate(u)
		case u.LastUpdateID < lastUpdateId:
			//fmt.Println("u.lastUpdateId", u.LastUpdateID, "is smaller than lastUpdateId", lastUpdateId)
			continue
		case u.FirstUpdateID > lastUpdateId:
			//fmt.Println("First Update ID", u.FirstUpdateID, "is larger than snapshot's lastUpdateId", lastUpdateId)
			stopInbound <- true
			close(outbound)
			return
		case u.FirstUpdateID <= lastUpdateId && u.LastUpdateID >= lastUpdateId:
			//fmt.Println("writing to queue: u.First", u.FirstUpdateID, "snapshot.Last", lastUpdateId, "u.Last", u.LastUpdateID)
			found = true
			outbound <- parseUpdate(u)
		}
	}
	//fmt.Println("writeValidUpdates: closing outbound channel", i)
	close(outbound) // inbound channel is closed, need to notify the Stream listener.
}
