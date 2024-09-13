package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("[USAGE]: ./client $IP $port")
		return
	}

	// Server IP and port
	server := os.Args[1] + ":" + os.Args[2]

	// Connect to the server
	conn, err := net.Dial("tcp", server)
	if err != nil {
		fmt.Println("Error connecting to the server:", err)
		return
	}
	defer conn.Close()

	// Goroutine to listen for messages from server
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// Send messages to the server
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		// Exit client if the message is "exit"
		if message == "exit" {
			break
		}

		conn.Write([]byte(message + "\n"))
	}
}
