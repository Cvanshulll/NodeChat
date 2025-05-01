package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/fatih/color"
)

func main() {
	// Connect to the server on localhost:9000
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		color.Red("❌ Could not connect to server: %v", err)
		return
	}
	color.Cyan("📡 Connected to server at localhost:9000")
	defer conn.Close()

	// Create a separate goroutine to read from the server
	go readFromServer(conn)

	// Send messages to the server
	scanner := bufio.NewScanner(os.Stdin)
	for {
		color.Set(color.FgGreen)
		fmt.Print("You: ")
		color.Unset()

		// Read user input and send to server
		if scanner.Scan() {
			text := scanner.Text() + "\n"
			conn.Write([]byte(text))
		}
	}
}

// readFromServer reads and prints messages from the server
func readFromServer(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Server disconnected.")
			os.Exit(0)
		}
		color.Magenta("Server: %s", msg)
	}
}
