package main

import (
	"os"
	"net/http"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// connected clients
var clients = make(map[*websocket.Conn]bool)

// broadcast channel
var broadcast = make(chan Message)

// Configure the upgrader
var upgrader = websocket.Upgrader{}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

// Define our message object
type Message struct {
	Email string `json:"email"`
	Username string `json:"username"`
	Message string `json:"message"`
}

func main() {
	// Create a simple file server
	// This is to display the UI
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/",fs)

	// These are run as go routines
	http.HandleFunc("/ws",handleConnections)

	// go routine handling the transmission of messages from server
	// to all the clients
	go handleMessages()

	log.Println("HTTP server started on :8000")
	err :=http.ListenAndServe(":8000",nil)
	if err !=nil {
		log.WithFields(log.Fields{
			"Error":    err,
		}).Fatal("Error in ListenAndServer!")
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrading Get request to Web Socket
	ws, err := upgrader.Upgrade(w,r, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"Error":    err,
		}).Fatal("Error in upgrading to WebSockets!")
	}

	defer ws.Close()

	clients[ws] = true

	for {
		var msg Message

		// Read in a new message as JSON and map to a message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":    err,
			}).Info("Error while reading JSON!")
			delete(clients,ws)
			break
		}

		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
			// Grab the next message from the broadcast channel
			msg := <-broadcast 
			// Send it out to every client that is currently connected
			for client := range clients {
					err := client.WriteJSON(msg)
					if err != nil {
							log.WithFields(log.Fields{
								"Error":    err,
							}).Info("Error while writing JSON!")
							client.Close()
							delete(clients, client)
					}
			}
	}
}
