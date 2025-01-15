package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Address     string
	Connections map[string]*User
	Messages    chan Message
	Mutex       sync.Mutex
}

type Message struct {
	SenderId       string
	SenderUsername string
	Content        string
	Timestamp      string
}

type User struct {
	UserId        string
	Username      string
	StoreFilePath string
	Conn          net.Conn
}

func (s *Server) Start() {
	listen, err := net.Listen("tcp", s.Address)
	if err != nil {
		fmt.Println("error in listen")
		panic(err)
	}

	defer listen.Close()
	fmt.Println("Server started on", s.Address)

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("error in accept")
			continue
		}

		go s.HandleConnection(conn)
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
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

	user := &User{
		UserId:        userId,
		Username:      username,
		StoreFilePath: storeFilePath,
		Conn:          conn,
	}

	s.Mutex.Lock()
	s.Connections[user.UserId] = user
	s.Mutex.Unlock()

	fmt.Printf("User connected: %s with file path: %s\n", username, storeFilePath)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("User disconnected: %s\n", username)
			break
		}

		messageContent := string(buffer[:n])
		if strings.HasPrefix(messageContent, "/sendfile") {
			parts := strings.SplitN(messageContent, " ", 3)
			if len(parts) < 3 {
				_, _ = conn.Write([]byte("Usage: /sendfile <userId> <filename>\n"))
				continue
			}
			recipientId := parts[1]
			filepath := parts[2]
			s.SendFile(userId, recipientId, filepath)
		} else if strings.HasPrefix(messageContent, "/FILE_REQUEST") {
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

			fileData := make([]byte, fileSize)
			_, err = io.ReadFull(conn, fileData)
			if err != nil {
				fmt.Printf("Error reading file data (expected %d bytes): %v\n", fileSize, err)
				return
			}
			s.HandleFileTransfer(conn, recipientId, fileName, int64(fileSize), fileData)
			continue
		} else {
			s.Messages <- Message{
				SenderId:       userId,
				SenderUsername: username,
				Content:        fmt.Sprintf("%s (Received at %s)", messageContent, time.Now().Format("15:04:05")),
				Timestamp:      time.Now().Format("15:04:05"),
			}
		}
	}

	s.Mutex.Lock()
	delete(s.Connections, userId)
	s.Mutex.Unlock()
}

func (s *Server) HandleFileTransfer(conn net.Conn, recipientId, fileName string, fileSize int64, fileData []byte) {
	recipient, exists := s.Connections[recipientId]
	if exists {
		_, err := recipient.Conn.Write([]byte(fmt.Sprintf("/FILE_RESPONSE %s %s %d\n", recipientId, fileName, fileSize)))
		if err != nil {
			fmt.Printf("Error sending file response to %s: %v\n", recipientId, err)
		}
		_, err = recipient.Conn.Write(fileData)
		if err != nil {
			fmt.Printf("Error sending file to %s: %v\n", recipientId, err)
		}
	} else {
		fmt.Printf("User %s not found\n", recipientId)
	}
}

func (s *Server) SendFile(senderId, recipientId, filePath string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	_, exists := s.Connections[recipientId]
	if !exists {
		fmt.Printf("User %s not found\n", recipientId)
		return
	}

	sender, exists := s.Connections[senderId]
	if !exists {
		fmt.Printf("User %s not found\n", senderId)
		return
	}

	_, err := sender.Conn.Write([]byte(fmt.Sprintf("/sendfile %s %s\n", recipientId, filePath)))
	if err != nil {
		fmt.Printf("Error sending file to %s: %v\n", recipientId, err)
	}
}

func (s *Server) Broadcast() {
	for message := range s.Messages {
		s.Mutex.Lock()
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
		s.Mutex.Unlock()
	}
}

func main() {
	server := Server{
		Address:     "localhost:8080",
		Connections: make(map[string]*User),
		Messages:    make(chan Message),
	}

	go server.Broadcast()
	server.Start()
}
