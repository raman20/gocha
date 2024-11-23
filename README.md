# Go Chat Server

This is a simple TCP chat server written in Go. It allows multiple clients to connect and communicate with each other. The server handles client connections, messages, and implements a basic banning mechanism for clients who send messages too frequently.

## Features

- Client connection management
- Message broadcasting to all connected clients
- Basic banning mechanism for clients who spam messages

## Requirements

- Go 1.16 or later

## Running the Server

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd <repository-directory>
   ```

2. Run the server:
   ```bash
   go run main.go
   ```

3. Connect to the server using a TCP client (e.g., `telnet` or a custom client).

## License

This project is licensed under the MIT License.