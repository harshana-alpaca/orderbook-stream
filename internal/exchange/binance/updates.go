package binance

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

func getUpdates(ctx context.Context, symbol string, stop chan bool) (<-chan update, error) {

	dialer := websocket.DefaultDialer
	urlStr := fmt.Sprintf("wss://data-stream.binance.vision/ws/%s@depth", strings.ToLower(symbol))
	conn, resp, err := dialer.Dial(urlStr, nil)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		conn.Close()
		return nil, fmt.Errorf("binance websocket could not be created")
	}
	if conn == nil {
		return nil, errors.New("conn is nil")
	}
	receiver := make(chan update, 100)
	//fmt.Println("Calling go routine to read from websocket")
	go func() {
		defer conn.Close()
		defer close(receiver)

		go func() {
			<-ctx.Done()
			//fmt.Println("context done, closing connection")
			conn.Close()
		}()

		for {
			var u update
			err := conn.ReadJSON(&u)
			if err != nil {
				fmt.Println(err)
				return
			}
			select {
			case receiver <- u:
				//fmt.Println("wrote update to the channel", u.FirstUpdateID, u.LastUpdateID)
			case <-ctx.Done():
				//fmt.Println("context done")
				return
			case <-stop:
				//fmt.Println("stop called")
				return
			default:
				//	fmt.Println("won't send update to the channel, something wrong")
			}
		}
	}()
	return receiver, nil

}
