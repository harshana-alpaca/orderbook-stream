package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestWebSocket(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	stopCh := make(chan bool)
	defer cancel()
	updates, err := getUpdates(ctx, "btcusdt", stopCh)
	if err != nil {
		t.Fatal(err)
		return
	}

	for update := range updates {
		r := rand.Int()
		json.NewEncoder(os.Stdout).Encode(update)
		fmt.Println("received update from the channel", update.FirstUpdateID, update.LastUpdateID, "sleeping", r%2000, "milli seconds")
		time.Sleep(time.Duration(r%2000) * time.Millisecond)
	}

}
