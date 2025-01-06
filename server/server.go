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
	UserId   string
	Username string
	Conn     net.Conn
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

	user := &User{
		UserId:   userId,
		Username: username,
		Conn:     conn,
	}

	s.Mutex.Lock()
	s.Connections[userId] = user
	s.Mutex.Unlock()

	fmt.Printf("User connected: %s\n", username)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("User disconnected: %s\n", username)
			break
		}

		messageContent := string(buffer[:n])
		s.Messages <- Message{
			SenderId:  username,
			Content:   messageContent,
			Timestamp: "",
		}
	}

	s.Mutex.Lock()
	delete(s.Connections, userId)
	s.Mutex.Unlock()
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
