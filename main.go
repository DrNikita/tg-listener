package main

import (
	"context"
	"fmt"

	"github.com/aliforever/go-tdlib"
	"github.com/aliforever/go-tdlib/config"
)

func main() {
	


	managerHandlers := tdlib.NewManagerHandlers().
		SetRawIncomingEventHandler(func(eventBytes []byte) {
			fmt.Println(string(eventBytes))
		})

	managerOptions := tdlib.NewManagerOptions().
		SetLogVerbosityLevel(6).
		SetLogPath("logs.txt")

	// Or you can pass nil for both handlers and options
	m := tdlib.NewManager(context.Background(), managerHandlers, managerOptions)

	// NewClientOptions
	cfg := config.New().
		SetFilesDirectory("./tdlib/tdlib-files").
		SetDatabaseDirectory("./tdlib/tdlib-db")

	h := tdlib.NewHandlers().
		SetRawIncomingEventHandler(func(eventBytes []byte) {
			fmt.Println(string(eventBytes))
		})

	apiID := int64(26790193)
	apiHash := "966d04d7114486eef7d5b9bb69f0948a"

	client := m.NewClient(apiID, apiHash, h, cfg, nil)

	err := client.ReceiveUpdates(context.Background())
	if err != nil {
		panic(err)
	}
}
