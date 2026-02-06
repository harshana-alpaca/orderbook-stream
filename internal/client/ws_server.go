package client

import (
	"MarketDataHub/internal/domain/orderbook"
	"MarketDataHub/internal/domain/orderbook/model"
	"MarketDataHub/internal/domain/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func StartServer(hub pubsub.Subscriber, service orderbook.Service) {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.StripSlashes)
	r.Get("/ws/streams", listenToHub(hub, service))
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func listenToHub(hub pubsub.Subscriber, service orderbook.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		streams := r.URL.Query()["streams"]
		fmt.Println("ws connection", streams)
		if len(streams) == 0 {
			_ = conn.WriteMessage(websocket.CloseMessage, []byte("No query params in the url"))
			fmt.Println("ws connection stopped")
			_ = conn.Close()
			return
		}
		symbols := strings.Split(streams[0], "/")

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		writeCh := make(chan []byte, 100)
		go readLoop(ctx, cancel, conn)
		go writeLoop(ctx, conn, writeCh)

		var wg sync.WaitGroup

		for _, symbol := range symbols {
			go getUpdates(symbol, hub, service, writeCh, &wg)
			wg.Add(1)
		}
		wg.Wait()

	}
}

func getUpdates(symbol string, hub pubsub.Subscriber, service orderbook.Service, returnCh chan<- []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	defer fmt.Println("calling wg.Done()")
	updateCh := make(chan model.Update, 100)
	rst := make(chan bool)
	fmt.Println("ws connection subs", symbol)
	err := hub.Subscribe(symbol, updateCh, rst)
	if err != nil {
		return
	}
	snapshot, _ := service.GetSnapshot(nil, symbol)
	expectedFirstUpdateId := snapshot.LastUpdateId + 1
	//wait for correct update thing
	synced := false
	for {
		select {
		case update := <-updateCh:
			if update.FirstUpdateId < expectedFirstUpdateId { //ignore
				fmt.Println("ignoring update", symbol, update.FirstUpdateId, update.LastUpdateId, expectedFirstUpdateId)
				continue
			}
			if !synced {
				if update.FirstUpdateId <= expectedFirstUpdateId && expectedFirstUpdateId < update.LastUpdateId {
					bytes, _ := json.Marshal(snapshot)
					returnCh <- bytes
					synced = true
				} else {
					fmt.Println("corrupted update", symbol, update.FirstUpdateId, update.LastUpdateId, expectedFirstUpdateId)

					_ = hub.Unsubscribe(symbol, updateCh)
					return
				}
			}
			if synced {
				bytes, _ := json.Marshal(update)
				returnCh <- bytes
			}
		}
	}

}

func readLoop(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) {
	defer cancel()
	conn.SetReadLimit(1024)
	//_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	//conn.SetPongHandler(func(string) error {
	//	err := conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	//	return err
	//})

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("received: %d %s %s\n", msgType, string(msg), err.Error())
			return
		}
		log.Printf("received: %d %s\n", msgType, string(msg))
	}
}

func writeLoop(ctx context.Context, conn *websocket.Conn, stream chan []byte) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	defer conn.Close()

	for {
		select {
		case msg, ok := <-stream:
			if !ok {
				log.Println("stream closed")
				return

			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println("error sending message")
				log.Println(err)
				return
			}
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, []byte("PING")); err != nil {
				log.Println(err)
				log.Println("error sending ping message")
				return
			}
		case <-ctx.Done():
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			log.Println("ctx done")
			return
		}
	}

}
