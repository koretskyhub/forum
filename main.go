package main

import (
	"forum/api"
	"log"
)

func main() {
	log.Println("Main started")
	if err := api.StartServer(); err != nil {
		log.Println(err)
	}
}
