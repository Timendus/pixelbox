package server

import (
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	bind             string
	router           *http.ServeMux
	connection       *Connection
	messageListeners []func([]byte)
}

var server Server

func init() {
	config := GetConfig()
	if len(config.Devices) == 0 {
		log.Fatal("No devices configured in config.json")
	}
	device := config.Devices[0]

	go func() {
		server.connection = NewConnection(device.Mac, device.Channel, func(msg []byte) {
			for _, listener := range server.messageListeners {
				listener(msg)
			}
		})
		err := server.connection.Connect()
		if err != nil {
			log.Println("Could not connect to the Divoom Timebox Evo:", err)
		} else {
			log.Println("Connected to Divoom Timebox Evo")
		}
	}()

	server = Server{
		bind:   config.Server.Host + ":" + strconv.Itoa(config.Server.Port),
		router: http.NewServeMux(),
	}
}

func RegisterMessageListener(listener func([]byte)) {
	server.messageListeners = append(server.messageListeners, listener)
}

func RegisterRouter(path string, router *http.ServeMux) {
	log.Println("Registering router for " + path)
	server.router.Handle(path+"/", http.StripPrefix(path, router))
}

func Static(serverPath, localPath string) {
	server.router.Handle(serverPath+"/", http.StripPrefix(serverPath, http.FileServer(http.Dir(localPath))))
}

func Root(path string) {
	server.router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, path, http.StatusFound)
	})
}

func Start() {
	log.Println("Starting server on " + server.bind)
	log.Fatal(http.ListenAndServe(server.bind, server.router))
}

func Stop() {
	if server.connection != nil {
		server.connection.Disconnect()
	}
}

func GetConnection() *Connection {
	return server.connection
}
