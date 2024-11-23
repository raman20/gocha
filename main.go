package main

import (
	"log"
	"net"
	"time"
)

// MessageType represents the type of message being sent.
type MessageType int

const (
	ClientConnected    MessageType = iota + 1 // Client connection event
	NewMessage                                // New message event
	ClientDisconnected                        // Client disconnection event
)

// Message represents a message sent by a client.
type Message struct {
	Type MessageType // Type of the message
	Conn net.Conn    // Connection of the author
	Text string      // Content of the message
}

// Client represents a connected client.
type Client struct {
	Conn        net.Conn  // Connection to the client
	LastMessage time.Time // Timestamp of the last message sent
	StrikeCount int
}

// server handles incoming messages and manages connected clients.
func server(messages chan Message) {
	clients := make(map[string]*Client)     // Map to track connected clients
	bannedClients := map[string]time.Time{} // Map to track banned clients
	for {
		msg := <-messages // Wait for a message
		switch msg.Type {
		case ClientConnected:

			addr := msg.Conn.RemoteAddr().(*net.TCPAddr)
			bannedAt, banned := bannedClients[addr.IP.String()]

			if banned {
				if time.Since(bannedAt).Seconds() >= 60.0 {
					delete(bannedClients, addr.IP.String())
					banned = false
				} else {
					msg.Conn.Write([]byte("you are banned\n"))
					msg.Conn.Close()
				}
			} else {
				log.Printf("Client %s connected", msg.Conn.RemoteAddr())
				clients[addr.IP.String()] = &Client{
					Conn:        msg.Conn,
					LastMessage: time.Now(),
				}
			}
		case NewMessage:
			addr := msg.Conn.RemoteAddr().(*net.TCPAddr)
			now := time.Now()
			author := clients[addr.IP.String()]
			if author != nil {
				if now.Sub(author.LastMessage).Seconds() >= 1.0 {
					author.LastMessage = now
					author.StrikeCount = 0
					log.Printf("Client %s sent message: %s", msg.Conn.RemoteAddr(), msg.Text)
					for _, client := range clients {
						if client.Conn.RemoteAddr().String() != msg.Conn.RemoteAddr().String() {
							_, err := client.Conn.Write([]byte(msg.Text))
							if err != nil {
								log.Printf("Could not send data to %s", client.Conn.RemoteAddr())
							}
						}
					}
				} else {
					author.StrikeCount += 1
					if author.StrikeCount >= 10 {
						bannedClients[addr.IP.String()] = now
						author.Conn.Write([]byte("you are banned\n"))
						author.Conn.Close()
					}
				}
			} else {
				msg.Conn.Close()
			}
		case ClientDisconnected:
			log.Printf("Client %s disconnected", msg.Conn.RemoteAddr())
			delete(clients, msg.Conn.RemoteAddr().String())
		}
	}
}

// client handles communication with a single client.
func handleClient(conn net.Conn, messages chan Message) {
	buffer := make([]byte, 64) // Buffer for reading messages
	for {
		n, err := conn.Read(buffer[:])
		if err != nil {
			log.Printf("Could not read from client %s", conn.RemoteAddr())
			conn.Close()
			messages <- Message{
				Type: ClientDisconnected,
				Conn: conn,
			}
			return
		}

		messages <- Message{
			Type: NewMessage,
			Text: string(buffer[0:n]),
			Conn: conn,
		}
	}
}

// main initializes the server and listens for incoming connections.
func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("ERROR: could not listen to port 8080: %s\n", err)
	}
	log.Printf("Listening to TCP connection on port 8080 ...\n")

	messages := make(chan Message) // Channel for message communication
	go server(messages)            // Start the server in a goroutine

	for {
		conn, err := ln.Accept() // Accept new connections
		if err != nil {
			log.Printf("Could not accept a connection: %s\n", err)
			continue
		}
		log.Printf("Accepted connection from %s", conn.RemoteAddr())

		messages <- Message{
			Type: ClientConnected,
			Conn: conn,
		}

		go handleClient(conn, messages) // Handle the client in a new goroutine
	}
}
