package main

import (
	"net"
	"sync"
)

type Server struct {
	Address     string
	Connections map[string]net.Conn
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
		panic(err)
	}

	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			panic(err)
		}

		go s.HandleConnection(conn)
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		panic(err)
	}
	userId := string(buffer[:n])

	s.Mutex.Lock()
	s.Connections[userId] = conn
	s.Mutex.Unlock()

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}

		s.Messages <- Message{
			SenderId:  userId,
			Content:   string(buffer[:n]),
			Timestamp: string(buffer[:n]),
		}
	}

	s.Mutex.Lock()
	delete(s.Connections, userId)
	s.Mutex.Unlock()
}

func (s *Server) Broadcast() {
	for message := range s.Messages {
		s.Mutex.Lock()
		for _, conn := range s.Connections {
			conn.Write([]byte(message.SenderId + ": " + message.Content))
		}
		s.Mutex.Unlock()
	}
}

func main() {
	server := Server{
		Address:     "localhost:8080",
		Connections: make(map[string]net.Conn),
		Messages:    make(chan Message),
	}

	go server.Broadcast()
	server.Start()
}
