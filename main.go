package main

import (
	"flag"
	"log"

	"github.com/mempool/mempool-cli/ui"
)

func main() {
	endpoint := flag.String("endpoint", "mempool.space", "The API endpoint")
	flag.Parse()

	gui, err := ui.New(*endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer gui.Close()

	if err := gui.Loop(); err != nil {
		log.Fatal(err)
	}
}
