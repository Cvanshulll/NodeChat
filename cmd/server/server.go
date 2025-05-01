package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/redis/go-redis/v9"
)

var clients = make(map[net.Conn]*redis.PubSub) // msp of client's connection to redis pubsub
var mu sync.Mutex
var rdb *redis.Client
var ctx = context.Background()

func main() {
	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
	})

	// Start the server on port 9000
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		color.Red("Failed to start server: %v", err)
		os.Exit(1)
	}
	color.Cyan("🚀 NodeChat server started on port 9000")

	// Continuously accept client connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			color.Red("Failed to accept connection: %v", err)
			continue
		}
		color.Green("New client connected: %s", conn.RemoteAddr().String())

		// Handle each client in a separate goroutine
		go handleConnection(conn)
	}

}

// handleConnection handles communication with a single client
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Ask client for the chatroom they want to join
	conn.Write([]byte("Enter chatroom name: "))
	reader := bufio.NewReader(conn)
	room, _ := reader.ReadString('\n')
	room = strings.TrimSpace(room)

	// Subscribe to the chatroom's Redis channel
	pubsub := rdb.Subscribe(ctx, room)
	clients[conn] = pubsub

	// Listen for messages from Redis
	go listenForMessages(conn, pubsub)

	// Handle incoming messages from the client
	Msgreader := bufio.NewReader(conn)
	for {
		msg, err := Msgreader.ReadString('\n')
		if err != nil {
			color.Yellow("Client %v disconnected", conn.RemoteAddr())
			removeClient(conn)
			return
		}
		msg = strings.TrimSpace(msg)
		// Publish the message to the Redis chatroom
		err = rdb.Publish(ctx, room, fmt.Sprintf("%v: %s", conn.RemoteAddr(), msg)).Err()
		if err != nil {
			color.Red("Failed to publish message: %v", err)
		}
	}
}

// Listen for messages from Redis and send them to the client
func listenForMessages(conn net.Conn, pubsub *redis.PubSub) {
	for msg := range pubsub.Channel() {
		conn.Write([]byte(msg.Payload + "\n"))
	}
}

// Remove a client from the map
func removeClient(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	delete(clients, conn)
}
