package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Maximum connections allowed
const maxConnections = 10

// Client structure
type Client struct {
	conn     net.Conn
	username string
}

var clients = make(map[net.Conn]Client)
var messages = []string{} // Store all previous messages
var mutex sync.Mutex      // Ensure safe access to shared data

// Broadcast message to all clients
func broadcastMessage(message string, exclude net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	messages = append(messages, message) // Store the message
	for conn := range clients {
		if conn != exclude { // Exclude sender from receiving its own message
			conn.Write([]byte(message + "\n"))
		}
	}
}

// Handle each client connection
func handleClient(conn net.Conn) {
	defer conn.Close()

	// Welcome the client and ask for the name
	conn.Write([]byte("Welcome to TCP-Chat!\n[ENTER YOUR NAME]: "))
	reader := bufio.NewReader(conn)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	// Ensure non-empty name
	if name == "" {
		conn.Write([]byte("Name cannot be empty.\n"))
		return
	}

	// Add client to the list
	mutex.Lock()
	clients[conn] = Client{conn: conn, username: name}
	mutex.Unlock()

	// Inform other clients of the new connection
	joinMessage := fmt.Sprintf("[%s][%s] has joined the chat", time.Now().Format("2006-01-02 15:04:05"), name)
	broadcastMessage(joinMessage, conn)

	// Send previous messages to the new client
	for _, msg := range messages {
		conn.Write([]byte(msg + "\n"))
	}

	// Listen for messages from the client
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			// Client disconnected
			break
		}
		message = strings.TrimSpace(message)
		if message != "" {
			formattedMessage := fmt.Sprintf("[%s][%s]: %s", time.Now().Format("2006-01-02 15:04:05"), name, message)
			broadcastMessage(formattedMessage, conn)
		}
	}

	// Client has left
	mutex.Lock()
	delete(clients, conn)
	mutex.Unlock()
	leaveMessage := fmt.Sprintf("[%s][%s] has left the chat", time.Now().Format("2006-01-02 15:04:05"), name)
	broadcastMessage(leaveMessage, conn)
}

// Start TCP server and listen for connections
func main() {
	port := "8989" // Default port
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
	defer listener.Close()

	log.Printf("Listening on port %s\n", port)

	for {
		if len(clients) >= maxConnections {
			log.Println("Maximum connections reached. Refusing new connections.")
			continue
		}

		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleClient(conn) // Handle each client in a separate goroutine
	}
}
