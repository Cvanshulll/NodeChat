package main

import (
    "bufio"
    "context"
    "fmt"
    "net"
    "strings"
    "time"

    "github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func main() {
    // Initialize Redis client
    rdb = redis.NewClient(&redis.Options{
        Addr: "localhost:6379", // Redis server address
    })

    _, err := rdb.Ping(ctx).Result()
    if err != nil {
        fmt.Printf("Failed to connect to Redis: %v\n", err)
        return
    }
    fmt.Println("Connected to Redis")

    // Start the server on port 9000
    listener, err := net.Listen("tcp", "0.0.0.0:9000") // Bind to all interfaces
    if err != nil {
        fmt.Printf("Failed to start server: %v\n", err)
        return
    }
    fmt.Println("🚀 Server started on port 9000 and accessible to all clients on the network")

    // Start broadcasting the server's IP
    go broadcastServerIP("9000")

    // Accept client connections
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Failed to accept connection: %v\n", err)
            continue
        }
        fmt.Printf("New client connected: %s\n", conn.RemoteAddr().String())
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    reader := bufio.NewReader(conn)

    // Ask for the username
    conn.Write([]byte("Enter your username: \n"))
    username, _ := reader.ReadString('\n')
    username = strings.TrimSpace(username)
    fmt.Printf("User '%s' connected from %s\n", username, conn.RemoteAddr().String())

    // Ask for chatroom name
    conn.Write([]byte("Enter the name of the chatroom you want to join: \n"))
    room, _ := reader.ReadString('\n')
    room = strings.TrimSpace(room)
    fmt.Printf("User '%s' joined chatroom: %s\n", username, room)

    // Subscribe to Redis channel
    pubsub := rdb.Subscribe(ctx, room)
    defer pubsub.Close()

    // Start listening for messages from Redis
    go func() {
        for msg := range pubsub.Channel() {
            // Check if the message was sent by this client
            if !strings.HasPrefix(msg.Payload, username+":") {
                conn.Write([]byte(msg.Payload + "\n"))
            }
        }
    }()

    // Handle incoming messages from the client
    for {
        msg, err := reader.ReadString('\n')
        if err != nil {
            fmt.Printf("User '%s' disconnected: %v\n", username, err)
            return
        }
        msg = strings.TrimSpace(msg)

        // Publish the message to the Redis chatroom with the username as a prefix
        rdb.Publish(ctx, room, fmt.Sprintf("%s: %s", username, msg))
    }
}

// Broadcast the server's IP address on the network
func broadcastServerIP(port string) {
    localIP := getLocalIP()
    if localIP == "" {
        fmt.Println("Failed to determine local IP for broadcasting.")
        return
    }

    addr := net.UDPAddr{
        IP:   net.IPv4bcast, // Broadcast address
        Port: 9001,          // Broadcast port
    }

    conn, err := net.DialUDP("udp", nil, &addr)
    if err != nil {
        fmt.Printf("Failed to set up UDP broadcast: %v\n", err)
        return
    }
    defer conn.Close()

    for {
        message := fmt.Sprintf("SERVER_IP:%s:%s", localIP, port)
        _, err := conn.Write([]byte(message))
        if err != nil {
            fmt.Printf("Failed to broadcast server IP: %v\n", err)
        }
        time.Sleep(5 * time.Second) // Broadcast every 5 seconds
    }
}

// Get the local IP address of the server
func getLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Printf("Error getting network interfaces: %v\n", err)
        return ""
    }

    for _, addr := range addrs {
        if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
            if ipNet.IP.To4() != nil {
                return ipNet.IP.String()
            }
        }
    }
    return ""
}