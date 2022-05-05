package main

import (
	"blockchaincore/cli"
	"log"
)

func OpenCLI() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Terminate program because of error: ", r)
		}
	}()

	c := cli.NewCLI()
	c.Run()
}

func main() {

	log.SetFlags(log.Lshortfile)
	OpenCLI()
}
