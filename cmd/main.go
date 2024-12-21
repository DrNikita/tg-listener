package main

import (
	"tg-listener/configs"
)

func main() {
	config, err := configs.MustConfig()
	if err != nil {
		if config != nil {

		}
	}
}
