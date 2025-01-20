package connection

import (
	"drizlink/server/interfaces"
	"fmt"
	"log"
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
	defer conn.Close()

	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading User Name:", err)
		return
	}
	username := string(buffer[:n])

	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading Store File Path:", err)
		return
	}
	storeFilePath := string(buffer[:n])

	userId := strconv.Itoa(rand.Intn(10000000))

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

	fmt.Printf("User connected: %s with ID: %s\n", username, userId)
	BroadcastMessage(fmt.Sprintf("User %s is now Online", username), server)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("User disconnected: %s\n", username)
			break
		}

		messageContent := string(buffer[:n])
		switch {
		case messageContent == "/exit":
			server.Mutex.Lock()
			user.IsOnline = false
			// delete(s.Connections, userId)
			server.Mutex.Unlock()
			BroadcastMessage(fmt.Sprintf("User %s is now offline", username), server)
			return
		case strings.HasPrefix(messageContent, "/FILE_REQUEST"):
			parts := strings.SplitN(messageContent, " ", 4)
			if len(parts) < 4 {
				_, _ = conn.Write([]byte("Usage: /FILE_REQUEST <userId> <filename> <filesize>\n"))
				continue
			}
			recipientId := parts[1]
			fileName := parts[2]
			fileSizeStr := strings.TrimSpace(parts[3])
			fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)

			if err != nil {
				fmt.Printf("Invalid file size '%s': %v\n", fileSizeStr, err)
				_, _ = conn.Write([]byte(fmt.Sprintf("Invalid file size: %s\n", fileSizeStr)))
				continue
			}
			HandleFileTransfer(server, conn, recipientId, fileName, int64(fileSize))
			continue
		case strings.HasPrefix(messageContent, "/status"):
			fmt.Println("Sending user list...")
			_, err = conn.Write([]byte("USERS:"))
			if err != nil {
				fmt.Println("Error sending user list:", err)
				return
			}
			for _, user := range server.Connections {
				if user.IsOnline {
					_, err = conn.Write([]byte(fmt.Sprintf("%s is online\n", user.Username)))
					if err != nil {
						fmt.Println("Error sending user list:", err)
						return
					}
					fmt.Printf("%s is online\n", user.Username)
				} else {
					_, err = conn.Write([]byte(fmt.Sprintf("%s is offline\n", user.Username)))
					if err != nil {
						fmt.Println("Error sending user list:", err)
						return
					}
					fmt.Printf("%s is offline\n", user.Username)
				}
			}
			continue
		default:
			server.Messages <- interfaces.Message{
				SenderId:       userId,
				SenderUsername: username,
				Content:        fmt.Sprintf("%s (Received at %s)", messageContent, time.Now().Format("15:04:05")),
				Timestamp:      time.Now().Format("15:04:05"),
			}
		}
	}

	server.Mutex.Lock()
	user.IsOnline = false
	// delete(s.Connections, userId)
	server.Mutex.Unlock()
	BroadcastMessage(fmt.Sprintf("User %s is now offline", username), server)
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

func Broadcast(server *interfaces.Server) {
	for message := range server.Messages {
		server.Mutex.Lock()
		var sb strings.Builder
		sb.WriteString("-> [")
		sb.WriteString(strings.TrimSpace(message.SenderUsername))
		sb.WriteString("] ")
		sb.WriteString(strings.TrimSpace(message.Content))
		formattedMsg := sb.String()

		log.SetFlags(0) // Disable default timestamp in log
		log.Print(formattedMsg)
		// for _, user := range s.Connections {
		// 	_, err := user.Conn.Write([]byte(formattedMsg))
		// 	if err != nil {
		// 		log.Printf("Error broadcasting message: %v", err)
		// 	}
		// }
		server.Mutex.Unlock()
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
