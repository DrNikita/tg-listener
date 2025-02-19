package main

import (
	"log"
	"tg-listener/internal/domen"
)

func main() {
	domenRepository, disconnect, destroy, err := domen.New()
	defer func() {
		disconnect()
		destroy()
	}()
	if err != nil {
		log.Fatal(err)
	}

	domenRepository.BackgroundListening()
}