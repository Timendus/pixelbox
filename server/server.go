package server

import (
	"log"
	"net/http"
	"os"
)

type Server struct {
	port             string
	router           *http.ServeMux
	connection       *Connection
	messageListeners []func([]byte)
}

var server Server

func init() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
		log.Println("No PORT environment variable detected. Defaulting to " + port)
	}

	go func() {
		server.connection = NewConnection("11:75:58:B1:B2:15", 1, func(msg []byte) {
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
		port:   ":" + port,
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
	log.Println("Starting server on localhost" + server.port)
	log.Fatal(http.ListenAndServe(server.port, server.router))
}

func Stop() {
	if server.connection != nil {
		server.connection.Disconnect()
	}
}

func GetConnection() *Connection {
	return server.connection
}
