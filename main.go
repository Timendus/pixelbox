package main

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"github.com/timendus/pixelbox/protocol"

	// Allow controllers to initialize themselves
	_ "github.com/timendus/pixelbox/controllers"

	"github.com/timendus/pixelbox/server"
)

var connection server.Connection

func main() {
	server.Static("/client", "./client")
	server.Root("/client")
	server.RegisterMessageListener(callback)
	defer server.Stop()
	server.Start()
}

func callback(message []byte) {
	commands, err := protocol.ParseIncoming(message)
	if err != nil {
		log.Println("Could not parse message:", err, message)
		return
	}
	log.Println("Parsed incoming message as:")
	for _, command := range commands {
		log.Println(" *", command)
	}
}
