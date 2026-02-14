package main

import (
	"embed"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"log"

	"github.com/timendus/pixelbox/controllers"
	"github.com/timendus/pixelbox/protocol"

	// Allow controllers to initialize themselves
	_ "github.com/timendus/pixelbox/controllers"

	"github.com/timendus/pixelbox/server"
)

//go:embed client
var client embed.FS

func main() {
	subDir, err := fs.Sub(client, "client")
	if err != nil {
		panic(err)
	}
	server.StaticFS("/client", subDir)
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
	for _, command := range commands {
		controllers.Events.Broadcast(command.String())
	}
}
