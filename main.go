package main

import (
	"log"

	"github.com/gchaincl/mempool/client"
	"github.com/gchaincl/mempool/ui"
)

func main() {
	c, err := client.New()
	if err != nil {
		log.Fatal(err)
	}

	gui, err := ui.New()
	if err != nil {
		log.Fatal(err)
	}
	defer gui.Close()

	go func() {
		for {
			resp, err := c.Read()
			if err != nil {
				log.Fatal(err)
			}

			gui.Render(resp)
		}
	}()

	if err := gui.Loop(); err != nil {
		log.Fatal(err)
	}
}
