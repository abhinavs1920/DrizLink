package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("error in dial")
		panic(err)
	}

	defer conn.Close()

	fmt.Println("Enter your userId: ")
	reader := bufio.NewReader(os.Stdin)
	userId, _ := reader.ReadString('\n')
	userId = strings.TrimSpace(userId)

	fmt.Println("Enter your username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	_, err = conn.Write([]byte(userId))
	if err != nil {
		fmt.Println("error in write userId")
		panic(err)
	}

	_, err = conn.Write([]byte(username))
	if err != nil {
		fmt.Println("error in write username")
		panic(err)
	}

	go func() {
		for {
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("error in read")
				return
			}
			fmt.Println(string(buffer[:n]))
		}
	}()

	fmt.Println("You can start typing your messages. Use '/sendfile <userId> <filename>' to send a file. Type 'exit' to quit.")

	for {
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		if message == "exit" {
			fmt.Println("Goodbye!")
			conn.Close()
			return
		}
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("error in write message", err)
			return
		}
	}
}
