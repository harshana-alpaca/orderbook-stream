package memory

import (
	"MarketDataHub/internal/domain/orderbook/model"
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
)

type Store struct {
	books map[string]*book
	mu    sync.RWMutex
}

type book struct {
	snapshot model.Snapshot
	lock     sync.RWMutex
}

func (s *Store) DeleteSnapshot(ctx context.Context, symbol string) error {
	delete(s.books, symbol)
	return nil
}

func NewStore() *Store {
	return &Store{
		books: make(map[string]*book),
	}
}

func (s *Store) ApplySnapshot(ctx context.Context, symbol string, snapshot model.Snapshot) (model.Update, error) {
	// this will be a blocking operation per channel.
	s.mu.Lock()

	if s.books[symbol] == nil {
		s.books[symbol] = &book{
			snapshot: snapshot,
		}
		defer s.mu.Unlock()
		return model.Update{}, nil
	}
	old := s.books[symbol]
	old.lock.Lock()
	s.mu.Unlock()
	defer s.books[symbol].lock.Unlock()
	update := calculateOrderBookDiff(old.snapshot, snapshot)
	s.books[symbol].snapshot = snapshot
	return update, nil
}

func (s *Store) ApplyUpdate(ctx context.Context, symbol string, update model.Update) error {
	s.mu.Lock()
	s.books[symbol].lock.Lock()
	defer s.books[symbol].lock.Unlock()
	s.mu.Unlock()
	currSnapshot := s.books[symbol]
	updatedSnapshot := insertUpdatesToOrderBook(currSnapshot.snapshot, update)
	s.books[symbol].snapshot = updatedSnapshot
	return nil
}

func (s *Store) GetSnapshot(ctx context.Context, symbol string) (model.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s, ok := s.books[symbol]; ok {
		return s.snapshot, nil
	}
	return model.Snapshot{}, errors.New("not found")
}

func insertUpdatesToOrderBook(book model.Snapshot, update model.Update) model.Snapshot {
	updatedBids := make([][2]string, 0, len(book.Bids)+len(update.Bids))
	i, j := 0, 0
	for i < len(book.Bids) && j < len(update.Bids) {
		diff, _ := comparePrice(book.Bids[i][0], update.Bids[j][0])
		if diff == 0 { //override the entry.
			if isFilled(update.Bids[j][1]) {
				//delete the entry. So skip.
				i++
				j++
				continue
			}
			//else
			updatedBids = append(updatedBids, update.Bids[j])
			i++
			j++
		}
		if diff < 0 { //we want the update's entry before books.
			if isFilled(update.Bids[j][1]) {
				// Don't need to do anything
				j++
				continue
			}
			updatedBids = append(updatedBids, update.Bids[j])
			j++
		}
		if diff > 0 { // need book's bid before update's
			updatedBids = append(updatedBids, book.Bids[i])
			i++
		}
	}
	updatedBids = append(updatedBids, book.Bids[i:]...)
	updatedBids = append(updatedBids, update.Bids[j:]...)
	book.Bids = updatedBids

	// process asks
	updatedAsks := make([][2]string, 0, len(book.Asks)+len(update.Asks))
	i, j = 0, 0
	for i < len(book.Asks) && j < len(update.Asks) {
		diff, _ := comparePrice(book.Asks[i][0], update.Asks[j][0])
		if diff == 0 { //override the entry.
			if isFilled(update.Asks[j][1]) {
				//delete the entry. So skip.
				i++
				j++
				continue
			}
			//else
			updatedAsks = append(updatedAsks, update.Asks[j])
			i++
			j++
		}
		if diff > 0 { //we want the update's entry before books. (lower ask first)
			if isFilled(update.Asks[j][1]) {
				// Don't need to do anything
				j++
				continue
			}
			updatedAsks = append(updatedAsks, update.Asks[j])
			j++
		}
		if diff < 0 { // need book's ask before update's
			updatedAsks = append(updatedAsks, book.Asks[i])
			i++
		}

	}
	updatedAsks = append(updatedAsks, book.Asks[i:]...)
	updatedAsks = append(updatedAsks, update.Asks[j:]...)
	book.Asks = updatedAsks
	book.LastUpdateId = update.LastUpdateId
	return book
}

func calculateOrderBookDiff(prev model.Snapshot, curr model.Snapshot) model.Update {
	update := model.Update{}
	update.Symbol = curr.Symbol
	update.LastUpdateId = curr.LastUpdateId
	update.FirstUpdateId = prev.LastUpdateId + 1

	bids := make([][2]string, 0, len(curr.Bids))
	asks := make([][2]string, 0, len(curr.Asks))
	// creating the bid's update
	// start with two pointers pointed to 0th index.
	// if both values are equal, check the price. if price is different, append the update's price.
	// if old price is less than current price, means we did not get the similar price in the updated one. add price:0 pair.
	// if current price is less than old price, means it wasn't in the books. add price:update.value pair.
	// iterate until end.
	i, j := 0, 0
	for i < len(prev.Asks) && j < len(curr.Asks) {
		// compare the values.
		diff, _ := comparePrice(prev.Asks[i][0], curr.Asks[j][0])
		switch {
		case diff == 0: // no diff in price
			if prev.Asks[i][1] != curr.Asks[j][1] {
				asks = append(asks, curr.Asks[j])
			}
			i++
			j++
		case diff < 0:
			ask := [2]string{prev.Asks[i][0], "0"}
			asks = append(asks, ask)
			i++
		default:
			// add the curr bid to the list.
			asks = append(asks, curr.Asks[j])
			j++
		}
	}
	for _, tup := range prev.Asks[i:] {
		asks = append(asks, [2]string{tup[0], "0"})
	}
	asks = append(asks, curr.Asks[j:]...)
	update.Asks = asks

	i, j = 0, 0
	for i < len(prev.Bids) && j < len(curr.Bids) {
		// compare the values.
		diff, _ := comparePrice(prev.Bids[i][0], curr.Bids[j][0])
		switch {
		case diff == 0: // no diff in price
			if prev.Bids[i][1] != curr.Bids[j][1] {
				bids = append(bids, curr.Bids[j])
			}
			i++
			j++
		case diff > 0:
			bid := [2]string{prev.Bids[i][0], "0"}
			bids = append(bids, bid)
			i++
		default:
			// add the curr bid to the list.
			bids = append(bids, curr.Bids[j])
			j++
		}
	}
	for _, tup := range prev.Bids[i:] {
		bids = append(bids, [2]string{tup[0], "0"})
	}
	bids = append(bids, curr.Bids[j:]...)
	update.Bids = bids
	return update
}

// return the value of price1-price2
func comparePrice(price1 string, price2 string) (int, error) {
	abs1 := strings.Replace(price1, ".", "", 1)
	abs2 := strings.Replace(price2, ".", "", 1)
	p1, err := strconv.Atoi(abs1)
	if err != nil {
		return 0, err
	}
	p2, err := strconv.Atoi(abs2)
	if err != nil {
		return 0, err
	}
	return p1 - p2, nil
}

func isFilled(qty string) bool {
	abs := strings.Replace(qty, ".", "", 1)
	if p, err := strconv.Atoi(abs); err == nil {
		return p == 0
	}
	return false
}
