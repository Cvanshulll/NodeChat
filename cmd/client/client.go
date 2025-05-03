package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"
    "time"
    "github.com/fatih/color"
)

func main() {
    fmt.Println("Attempting to discover server on the network...")
    cyan := color.New(color.FgCyan).SprintFunc()
    serverIP, serverPort := discoverServerIP()

    if serverIP == "" || serverPort == "" {
        // Fallback to manual input if discovery fails
        reader := bufio.NewReader(os.Stdin)
        fmt.Print("Enter server IP (default: use local Wi-Fi network): ")
        serverIP, _ = reader.ReadString('\n')
        serverIP = strings.TrimSpace(serverIP)

        if serverIP == "" {
            fmt.Println("Server IP is required.")
            return
        }

        fmt.Print("Enter server port (default: 9000): ")
        serverPort, _ = reader.ReadString('\n')
        serverPort = strings.TrimSpace(serverPort)
        if serverPort == "" {
            serverPort = "9000" // Default port
        }
    }

    serverAddress := fmt.Sprintf("%s:%s", "localhost", serverPort)
    fmt.Printf("Connecting to server at %s\n", serverAddress)

    // Connect to the server
    conn, err := net.Dial("tcp", serverAddress)
    if err != nil {
        fmt.Printf("Failed to connect to server: %v\n", err)
        return
    }
    defer conn.Close()
    flag := false

    // Read messages from the server
    go func() {
        for {
            if flag == true {
                return
            }
            message, err := bufio.NewReader(conn).ReadString('\n')
            if err != nil {
                fmt.Printf("Error reading from server: %v\n", err)
                return
            }
            fmt.Printf("\r%s\n", strings.TrimSpace(message)) 
            fmt.Print(cyan("you: "))                             
        }
    }()

    // Send messages to the server
    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print(cyan("you: "))
        text, _ := reader.ReadString('\n')
        conn.Write([]byte(strings.TrimSpace(text) + "\n"))
        if strings.TrimSpace(text) == "exit" {
            fmt.Println("Exiting...")
            flag = true
            return
        }
    }
}

// Discover the server's IP address via UDP broadcast
func discoverServerIP() (string, string) {
    addr := net.UDPAddr{
        IP:   net.IPv4zero, // Listen on all interfaces
        Port: 9001,         // Broadcast port
    }

    conn, err := net.ListenUDP("udp", &addr)
    if err != nil {
        fmt.Printf("Failed to listen for server broadcast: %v\n", err)
        return "", ""
    }
    defer conn.Close()

    buffer := make([]byte, 1024)
    conn.SetReadDeadline(time.Now().Add(10 * time.Second)) // Timeout after 10 seconds

    n, _, err := conn.ReadFromUDP(buffer)
    if err != nil {
        fmt.Println("No broadcast received. Please enter the server IP manually.")
        return "", ""
    }

    message := strings.TrimSpace(string(buffer[:n]))
    if strings.HasPrefix(message, "SERVER_IP:") {
        parts := strings.Split(message, ":")
        if len(parts) == 3 {
            return parts[1], parts[2] // Return IP and port
        }
    }
    return "", ""
}