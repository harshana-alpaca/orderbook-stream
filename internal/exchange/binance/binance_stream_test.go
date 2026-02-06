package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestBinanceStream(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	streamer := NewStreamer()
	_, updateCh, err := streamer.Stream(ctx, "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	var lastUpdateId uint64
	fmt.Println(lastUpdateId)
	for update := range updateCh {
		if update.FirstUpdateId == lastUpdateId+1 || lastUpdateId == 0 {
			//valid.
			_ = json.NewEncoder(os.Stdout).Encode(update)
			fmt.Println(lastUpdateId, update.FirstUpdateId, update.LastUpdateId)
			lastUpdateId = update.LastUpdateId
		} else {
			t.Fatal("last update id not matching")
			return
		}
	}
}
