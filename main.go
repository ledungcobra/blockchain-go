package main

import (
	"blockchaincore/blockchain"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Terminate program because of error: ", r)
		}
	}()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	cli := blockchain.NewCLI()
	cli.Run()
}
