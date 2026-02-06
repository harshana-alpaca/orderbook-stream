package client

import "github.com/gorilla/websocket"

type Client struct {
	symbols    []string
	connection *websocket.Conn
}

func NewClient(symbols []string, connection *websocket.Conn) *Client {
	return &Client{}
}
