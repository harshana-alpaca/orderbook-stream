package cmd

import (
	"MarketDataHub/internal/client"
	"MarketDataHub/internal/domain/orderbook"
	"MarketDataHub/internal/domain/pubsub"
	"MarketDataHub/internal/exchange"
	"MarketDataHub/internal/exchange/binance"
	"MarketDataHub/internal/store/memory"
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "market-data-server",
	Short: "market-data-server receives data from exchanges and broadcast it via websocket connections.",
	Run: func(cmd *cobra.Command, args []string) {
		var connector exchange.Streamer = binance.NewStreamer()
		var memoryStore orderbook.Keeper = memory.NewStore()
		hub := pubsub.NewHub()
		var service = orderbook.NewService(connector, memoryStore, hub, []string{"btcusdt", "bnbusdt"})
		_ = service.Start(context.Background())
		client.StartServer(hub, *service)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
