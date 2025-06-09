package main

import "github.com/pradiptaagus/go-chat-app/internal/app"

func main() {
	client := app.NewClient()
	client.Run()
}
