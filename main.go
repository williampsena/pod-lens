package main

import (
	"log"

	"github.com/williampsena/pod-lens/internal/server"
)

func main() {
	if err := server.RunAndServer(); err != nil {
		log.Fatal(err)
	}
}
