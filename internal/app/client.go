package app

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pradiptaagus/go-chat-app/utils"
)

type Client struct {
	Addr url.URL
}

func (client *Client) Run() {
	con, res, err := websocket.DefaultDialer.Dial(client.Addr.String(), nil)
	utils.PanicIfError(err)
	log.Println("dial:", res)
	defer con.Close()

	ticker := time.NewTicker(time.Second)
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			_, msg, err := con.ReadMessage()
			if err != nil {
				log.Println("read:", err)
			}
			log.Printf("Received message: %s\n", msg)
		}
	}()

	for {
		select {
		case <-signalChan:
			fmt.Println("Interrupted")
			err := con.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				fmt.Println("write close:", err)
			}
			return
		case t := <-ticker.C:
			fmt.Println("Tick at", t)
			message := []byte(t.String())
			err := con.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println("write:", err)
			}
		}
	}
}

func NewClient() *Client {
	client := Client{
		Addr: url.URL{
			Scheme: "ws",
			Host:   "localhost:8080",
			Path:   "/chat",
		},
	}
	return &client
}
