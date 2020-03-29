package main

import (
	"log"

	"github.com/mempool/mempool-cli/ui"
)

func main() {
	gui, err := ui.New()
	if err != nil {
		log.Fatal(err)
	}
	defer gui.Close()

	if err := gui.Loop(); err != nil {
		log.Fatal(err)
	}
}
