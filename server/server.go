package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	Address     string
	Connections map[string]*User
	Messages    chan Message
	Mutex       sync.Mutex
}

type Message struct {
	SenderId  string
	Content   string
	Timestamp string
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
		fmt.Println("Error reading User ID:", err)
		return
	}
	userId := string(buffer[:n])

	n, err = conn.Read(buffer)
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

	user := &User{
		UserId:        userId,
		Username:      username,
		StoreFilePath: storeFilePath,
		Conn:          conn,
	}

	s.Mutex.Lock()
	s.Connections[userId] = user
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
		} else {
			s.Messages <- Message{
				SenderId:  username,
				Content:   messageContent,
				Timestamp: "",
			}
		}
	}

	s.Mutex.Lock()
	delete(s.Connections, userId)
	s.Mutex.Unlock()
}

func (s *Server) SendFile(senderId, recipientId, filePath string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	recipient, exists := s.Connections[recipientId]
	if !exists {
		fmt.Printf("User %s not found\n", recipientId)
		return
	}

	_, err := recipient.Conn.Write([]byte(fmt.Sprintf("/sendfile %s %s\n", senderId, filePath)))
	if err != nil {
		fmt.Printf("Error sending file to %s: %v\n", recipientId, err)
	}
}

func (s *Server) Broadcast() {
	for message := range s.Messages {
		s.Mutex.Lock()
		// timestamp := time.Now().Format("15:04:05")
		var sb strings.Builder
		// sb.WriteString(timestamp)
		sb.WriteString(" [")
		sb.WriteString(strings.TrimSpace(message.SenderId))
		sb.WriteString("]: ")
		sb.WriteString(strings.TrimSpace(message.Content))
		sb.WriteString("\n")
		formattedMsg := sb.String()

		log.Print(formattedMsg)
		for _, user := range s.Connections {
			_, err := user.Conn.Write([]byte(formattedMsg))
			if err != nil {
				log.Printf("Error broadcasting message: %v", err)
			}
		}
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
