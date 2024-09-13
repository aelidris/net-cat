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
	name     string
	outgoing chan string
}

type Server struct {
	clients    map[*Client]bool
	broadcast  chan string
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mutex.Lock()
			s.clients[client] = true
			s.mutex.Unlock()
			s.broadcastMessage(fmt.Sprintf("%s has joined our chat...", client.name))
			s.sendPreviousMessages(client)
		case client := <-s.unregister:
			s.mutex.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.outgoing)
			}
			s.mutex.Unlock()
			s.broadcastMessage(fmt.Sprintf("%s has left our chat...", client.name))
		case message := <-s.broadcast:
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

func (s *Server) broadcastMessage(message string) {
	s.broadcast <- message
}

func (s *Server) sendPreviousMessages(client *Client) {
	// Implement logic to send previous messages to the new client
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	client := &Client{
		conn:     conn,
		outgoing: make(chan string),
	}

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
		" |    `.       | `' \\Zq\n" +
		"_)      \\.___.,|     .'\n" +
		"\\____   )MMMMMP|   .'\n" +
		"     `-'       `--'\n" +
		"[ENTER YOUR NAME]: "
	conn.Write([]byte(welcomeMessage))

	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading client name:", err)
		return
	}
	client.name = strings.TrimSpace(name)

	if client.name == "" {
		conn.Write([]byte("Name cannot be empty.\n"))
		return
	}

	s.register <- client

	go s.readMessages(client)
	s.writeMessages(client)
}

func (s *Server) readMessages(client *Client) {
	reader := bufio.NewReader(client.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		message = strings.TrimSpace(message)
		if message != "" {
			formattedMessage := fmt.Sprintf("[%s][%s]:%s", time.Now().Format("2006-01-02 15:04:05"), client.name, message)
			s.broadcastMessage(formattedMessage)
		}
	}
	s.unregister <- client
}

func (s *Server) writeMessages(client *Client) {
	for message := range client.outgoing {
		_, err := client.conn.Write([]byte(message + "\n"))
		if err != nil {
			break
		}
	}
}

func main() {
	port := "8990"
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
		if len(server.clients) >= 10 {
			conn.Write([]byte("Server is full. Try again later.\n"))
			conn.Close()
			continue
		}
		go server.handleClient(conn)
	}
}
