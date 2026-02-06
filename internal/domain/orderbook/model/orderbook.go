package model

type Snapshot struct {
	Symbol       string
	Bids         [][2]string
	Asks         [][2]string
	isValid      bool
	LastUpdateId uint64
}

type Update struct {
	Symbol        string
	Bids          [][2]string
	Asks          [][2]string
	LastUpdateId  uint64
	FirstUpdateId uint64
}
