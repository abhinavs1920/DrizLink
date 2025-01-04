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
