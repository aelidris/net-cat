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

type Client struct {
	conn     net.Conn
	username string
	outgoing chan string
}

type Server struct {
	clients    map[*Client]bool
	broadcast  chan string
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
}

// Maximum connections allowed
const maxConnections = 2

var (
	clients  = make(map[net.Conn]Client)
	messages = []string{} // Store all previous messages
	mutex    sync.Mutex   // Ensure safe access to shared data
)

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

	// Send welcome message and get client name
	welcomeMessage := "Welcome to TCP-Chat!\n" +
		"         _nnnn_\n" +
		"        dGGGGMMb\n" +
		"       @p~qp~~qMb\n" +
		"       M|@||@) M|\n" +
		"       @,----.JM|\n" +
		"      JS^\\__/  qKL\n" +
		"     dZP        qKRb\n" +
		"    dZP          qKKb\n" +
		"   fZP            SMMb\n" +
		"   HZM            MMMM\n" +
		"   FqM            MMMM\n" +
		" __| \".        |\\dS\"qML\n" +
		" |    .       | ' \\Zq\n" +
		"_)      \\.___.,|     .'\n" +
		"\\____   )MMMMMP|   .'\n" +
		"     -'       --'\n" +
		"[ENTER YOUR NAME]: "
	conn.Write([]byte(welcomeMessage))
	reader := bufio.NewReader(conn)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	// Ensure non-empty name
	if name == "" {
		conn.Write([]byte("Name cannot be empty.\n"))
		return
	}

	// Check if the server is full
	mutex.Lock()
	if len(clients) >= maxConnections {
		conn.Write([]byte("Server is full. Try again later.\n"))
		mutex.Unlock()
		conn.Close() // Close the connection after sending the message
		return
	}

	// Add client to the list
	clients[conn] = Client{conn: conn, username: name}
	mutex.Unlock()

	// Inform other clients of the new connection
	joinMessage := fmt.Sprintf("%s has joined our chat...", name)
	broadcastMessage(joinMessage, conn)

	// Send previous messages to the new client
	for _, msg := range messages {
		if msg == joinMessage {
			continue
		}
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
	leaveMessage := fmt.Sprintf("%s has left our chat...", name)
	broadcastMessage(leaveMessage, conn)
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mutex.Lock()
			s.clients[client] = true
			s.mutex.Unlock()
			log.Printf("New client registered: %s\n", client.username)
			s.broadcast <- fmt.Sprintf("%s has joined our chat...", client.username)
		case client := <-s.unregister:
			s.mutex.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.outgoing)
			}
			s.mutex.Unlock()
			log.Printf("Client unregistered: %s\n", client.username)
			s.broadcast <- fmt.Sprintf("%s has left our chat...", client.username)
		case message := <-s.broadcast:
			log.Printf("Broadcasting message: %s\n", message)
			s.mutex.Lock()
			for client := range s.clients {
				select {
				case client.outgoing <- message:
				default:
					close(client.outgoing)
					delete(s.clients, client)
				}
			}
			s.mutex.Unlock()
		}
	}
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func main() {
	port := "8989"
	if len(os.Args) > 1 {
		if len(os.Args) != 2 {
			fmt.Println("[USAGE]: ./TCPChat $port")
			return
		}
		port = os.Args[1]
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	fmt.Println("Listening on the port :" + port)

	server := NewServer()
	go server.Run()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleClient(conn)
	}
}
