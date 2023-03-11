package main

import (
	"log"
	"time"

	"github.com/ferama/crauti/cmd"
	"github.com/ferama/crauti/pkg/conf"
)

func main() {
	go func() {
		for {
			c, _ := conf.Dump()
			log.Printf("current conf:\n\n%s\n", c)
			time.Sleep(3 * time.Second)
		}
	}()

	cmd.Execute()
}
