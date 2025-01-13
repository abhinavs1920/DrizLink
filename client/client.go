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

	_, err = conn.Write([]byte(userId))
	if err != nil {
		fmt.Println("error in write userId")
		panic(err)
	}

	fmt.Println("Enter your username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	_, err = conn.Write([]byte(username))
	if err != nil {
		fmt.Println("error in write username")
		panic(err)
	}

	fmt.Println("Enter the file path to capture:")
	storeFilePath, _ := reader.ReadString('\n')
	storeFilePath = strings.TrimSpace(storeFilePath)

	_, err = conn.Write([]byte(storeFilePath))
	if err != nil {
		fmt.Println("error in write storeFilePath")
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
		} else if strings.HasPrefix(message, "/sendfile") {
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println("Invalid arguments. Use: /sendfile <userId> <filename>")
				continue
			}
			recipientId := args[1]
			filePath := args[2]
			handleSendFile(conn, recipientId, filePath)
			continue
		}
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("error in write message", err)
			return
		}
	}
}

func handleSendFile(conn net.Conn, recipientId, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error in open file", err)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("error in stat file", err)
		return
	}

	fileSize := fileInfo.Size()
	fileName := fileInfo.Name()
	buffer := make([]byte, fileSize)

	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println("error in read file", err)
		return
	}

	_, err = conn.Write([]byte(fmt.Sprintf("/sendfile %s %s %d\n", recipientId, fileName, fileSize)))
	if err != nil {
		fmt.Println("error in write sendfile", err)
		return
	}

	_, err = conn.Write(buffer)
	if err != nil {
		fmt.Println("error in write buffer", err)
		return
	}
}
