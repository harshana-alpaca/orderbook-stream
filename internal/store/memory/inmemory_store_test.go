package memory

import (
	"MarketDataHub/internal/domain/orderbook/model"
	"context"
	"fmt"
	"testing"
)

func TestReplaceSnapshot(t *testing.T) {
	prev := model.Snapshot{
		LastUpdateId: 100,
		Symbol:       "BTCUSDT",
		Bids: [][2]string{
			{"8", "1000"},
			{"7", "2000"},
			{"6", "3000"},
			{"5", "4000"},
		},
		Asks: [][2]string{
			{"10", "1000"},
			{"11", "2000"},
			{"12", "3000"},
			{"13", "4000"},
		},
	}

	curr := model.Snapshot{
		LastUpdateId: 200,
		Symbol:       "BTCUSDT",
		Bids: [][2]string{
			{"9", "1000"},
			{"8", "1000"},
			{"7", "3000"},
			{"6", "4000"},
			{"4", "4000"},
		},
		Asks: [][2]string{
			{"11", "1000"},
			{"12", "2000"},
			{"14", "3000"},
			{"15", "4000"},
		},
	}
	expectedDiff := model.Update{
		LastUpdateId: 200,
		Symbol:       "BTCUSDT",

		Bids: [][2]string{
			{"9", "1000"},
			{"7", "3000"},
			{"6", "4000"},
			{"5", "0"},
			{"4", "4000"},
		},
		Asks: [][2]string{
			{"10", "0"},
			{"11", "1000"},
			{"12", "2000"},
			{"13", "0"},
			{"14", "3000"},
			{"15", "4000"},
		},
	}
	diff := calculateOrderBookDiff(prev, curr)
	switch {
	case diff.Symbol != expectedDiff.Symbol:
		t.FailNow()
	case diff.LastUpdateId != expectedDiff.LastUpdateId:
		t.FailNow()
	}

	for i, bid := range diff.Bids {
		if bid[0] != expectedDiff.Bids[i][0] {
			t.FailNow()
		}
	}
	for i, ask := range diff.Asks {
		if ask[0] != expectedDiff.Asks[i][0] {
			t.FailNow()
		}
	}
	fmt.Println(diff.Bids, diff.Asks)

}

func TestUpdateSnapshot(t *testing.T) {
	snapshot := model.Snapshot{
		LastUpdateId: 100,
		Symbol:       "BTCUSDT",
		Bids: [][2]string{
			{"8", "1000"},
			{"7", "2000"},
			{"6", "3000"},
			{"5", "4000"},
		},
		Asks: [][2]string{
			{"10", "1000"},
			{"11", "2000"},
			{"12", "3000"},
			{"13", "4000"},
		},
	}

	update := model.Update{
		FirstUpdateId: 101,
		LastUpdateId:  200,
		Symbol:        "BTCUSDT",
		Bids: [][2]string{
			{"9", "1000"},
			{"7", "3000"},
			{"6", "4000"},
			{"5", "0"},
			{"4", "4000"},
		},
		Asks: [][2]string{
			{"10", "0"},
			{"11", "1000"},
			{"12", "2000"},
			{"13", "0"},
			{"14", "3000"},
			{"15", "4000"},
		},
	}
	store := NewStore()
	store.ApplySnapshot(context.Background(), "BTCUSDT", snapshot)
	_ = store.ApplyUpdate(context.Background(), "BTCUSDT", update)
	updatedBook, _ := store.GetSnapshot(context.Background(), "BTCUSDT")
	fmt.Println(updatedBook.Bids)
	fmt.Println(updatedBook.Asks)
	fmt.Println(updatedBook.LastUpdateId)
}
