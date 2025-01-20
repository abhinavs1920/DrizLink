package connection

import (
	"drizlink/server/interfaces"
	"drizlink/server/internal/encryption"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func Connect(address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return nil, err
	}
	return listener, nil
}

func Close(conn net.Conn) {
	conn.Close()
}

func Start(server *interfaces.Server) {
	listen, err := net.Listen("tcp", server.Address)
	if err != nil {
		fmt.Println("error in listen")
		panic(err)
	}

	defer listen.Close()
	fmt.Println("Server started on", server.Address)

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("error in accept")
			continue
		}

		go HandleConnection(conn, server)
	}
}

func HandleConnection(conn net.Conn, server *interfaces.Server) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("error in read username")
		return
	}
	encryptedUsername := string(buffer[:n])
	fmt.Println(encryptedUsername)
	// Decrypt username
	username, err := encryption.DecryptMessage(encryptedUsername)
	if err != nil {
		fmt.Println("error decrypting username:", err)
		return
	}

	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Println("error in read storeFilePath")
		return
	}
	encryptedStoreFilePath := string(buffer[:n])

	// Decrypt store file path
	storeFilePath, err := encryption.DecryptMessage(encryptedStoreFilePath)
	if err != nil {
		fmt.Println("error decrypting storeFilePath:", err)
		return
	}

	userId := generateUserId()

	user := &interfaces.User{
		UserId:        userId,
		Username:      username,
		StoreFilePath: storeFilePath,
		Conn:          conn,
		IsOnline:      true,
	}

	server.Mutex.Lock()
	server.Connections[user.UserId] = user
	server.Mutex.Unlock()

	// Encrypt and broadcast welcome message
	welcomeMsg := fmt.Sprintf("User %s has joined the chat", username)
	encryptedMsg, err := encryption.EncryptMessage(welcomeMsg)
	if err == nil {
		BroadcastMessage(encryptedMsg, server)
	}

	fmt.Printf("New user connected: %s (ID: %s)\n", username, userId)

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("User disconnected: %s\n", username)
			server.Mutex.Lock()
			user.IsOnline = false
			server.Mutex.Unlock()
			// Encrypt and broadcast offline message
			offlineMsg := fmt.Sprintf("User %s is now offline", username)
			encryptedOffline, err := encryption.EncryptMessage(offlineMsg)
			if err == nil {
				BroadcastMessage(encryptedOffline, server)
			}
			return
		}

		messageContent := string(buffer[:n])

		// Try to decrypt the message if it's not a command
		if !strings.HasPrefix(messageContent, "/") && messageContent != "PONG\n" {
			decryptedMsg, err := encryption.DecryptMessage(messageContent)
			if err == nil {
				messageContent = decryptedMsg
			}
		}

		switch {
		case messageContent == "/exit":
			server.Mutex.Lock()
			user.IsOnline = false
			server.Mutex.Unlock()
			// Encrypt and broadcast offline message
			offlineMsg := fmt.Sprintf("User %s is now offline", username)
			encryptedOffline, err := encryption.EncryptMessage(offlineMsg)
			if err == nil {
				BroadcastMessage(encryptedOffline, server)
			}
			return
		case strings.HasPrefix(messageContent, "/FILE_REQUEST"):
			args := strings.SplitN(messageContent, " ", 4)
			if len(args) != 4 {
				fmt.Println("Invalid arguments. Use: /FILE_REQUEST <userId> <filename> <fileSize>")
				continue
			}
			recipientId := args[1]
			fileName := args[2]
			fileSizeStr := strings.TrimSpace(args[3])
			fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
			if err != nil {
				fmt.Println("Invalid fileSize. Use: /FILE_REQUEST <userId> <filename> <fileSize>")
				continue
			}

			HandleFileTransfer(server, conn, recipientId, fileName, fileSize)
			continue
		case messageContent == "PONG\n":
			continue
		case strings.HasPrefix(messageContent, "/status"):
			_, err = conn.Write([]byte("USERS:"))
			if err != nil {
				fmt.Println("Error sending user list header:", err)
				continue
			}
			for _, user := range server.Connections {
				if user.IsOnline {
					// Encrypt user status message
					statusMsg := fmt.Sprintf("%s is online\n", user.Username)
					encryptedStatus, err := encryption.EncryptMessage(statusMsg)
					if err == nil {
						_, err = conn.Write([]byte(encryptedStatus))
						if err != nil {
							fmt.Println("Error sending user list:", err)
							continue
						}
					}
				}
			}
			continue
		default:
			// Encrypt and broadcast regular messages
			encryptedMsg, err := encryption.EncryptMessage(fmt.Sprintf("%s: %s", username, messageContent))
			if err == nil {
				BroadcastMessage(encryptedMsg, server)
			}
		}
	}
}

func generateUserId() string {
	return strconv.Itoa(rand.Intn(10000000))
}

func BroadcastMessage(content string, server *interfaces.Server) {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()
	for _, user := range server.Connections {
		if user.IsOnline {
			_, _ = user.Conn.Write([]byte(content + "\n"))
		}
	}
}

func StartHeartBeat(interval time.Duration, server *interfaces.Server) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			server.Mutex.Lock()
			for _, user := range server.Connections {
				if user.IsOnline {
					_, err := user.Conn.Write([]byte("PING\n"))
					if err != nil {
						fmt.Printf("User disconnected: %s\n", user.Username)
						user.IsOnline = false
						BroadcastMessage(fmt.Sprintf("User %s is now offline", user.Username), server)
					}
				}
			}
			server.Mutex.Unlock()
		}
	}()
}
