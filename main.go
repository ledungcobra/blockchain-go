package main

import (
	"blockchaincore/blockchain"
	"blockchaincore/web"
	"log"
)

func OpenCLI() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Terminate program because of error: ", r)
		}
	}()

	cli := blockchain.NewCLI()
	cli.Run()
}

func main() {
	web.StartWebServer("8080")

}
