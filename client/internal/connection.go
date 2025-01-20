package connection

import (
	"bufio"
	"drizlink/client/internal/encryption"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func Connect(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func Close(conn net.Conn) {
	conn.Close()
}

func UserInput(attribute string, conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter your " + attribute + ": ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// Encrypt user input before sending
	encryptedInput, err := encryption.EncryptMessage(input)
	if err != nil {
		fmt.Println("error encrypting " + attribute)
		panic(err)
	}

	_, err = conn.Write([]byte(encryptedInput))
	if err != nil {
		fmt.Println("error in write " + attribute)
		panic(err)
	}
}

func ReadLoop(conn net.Conn) {
	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("error in read")
			return
		}
		message := string(buffer[:n])

		// Try to decrypt the message if it's not a special command
		if !strings.HasPrefix(message, "/") && !strings.HasPrefix(message, "PING") {
			decryptedMsg, err := encryption.DecryptMessage(message)
			if err == nil {
				message = decryptedMsg
			}
		}

		switch {
		case strings.HasPrefix(message, "/FILE_RESPONSE"):
			fmt.Println("File transfer response received")
			args := strings.SplitN(message, " ", 5)
			if len(args) != 5 {
				fmt.Println("Invalid arguments. Use: /FILE_RESPONSE <userId> <filename> <fileSize> <storeFilePath>")
				continue
			}
			recipientId := args[1]
			fileName := args[2]
			fileSizeStr := strings.TrimSpace(args[3])
			fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
			storeFilePath := args[4]
			if err != nil {
				fmt.Println("Invalid fileSize. Use: /FILE_RESPONSE <userId> <filename> <fileSize> <storeFilePath>")
				continue
			}

			HandleFileTransfer(conn, recipientId, fileName, int64(fileSize), storeFilePath)
			continue
		case strings.HasPrefix(message, "PING"):
			_, err = conn.Write([]byte("PONG\n"))
			if err != nil {
				fmt.Println("Error responding to heartbeat: ", err)
				continue
			}
		case message == "USERS:":
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("error in read message", err)
				continue
			}
			userList := string(buffer[:n])
			// Try to decrypt the user list
			decryptedList, err := encryption.DecryptMessage(userList)
			if err == nil {
				userList = decryptedList
			}
			fmt.Println(userList)
			continue
		default:
			fmt.Println(message)
		}
	}
}

func WriteLoop(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		switch {
		case message == "exit":
			fmt.Println("Goodbye!")
			conn.Close()
			return
		case strings.HasPrefix(message, "/sendfile"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println("Invalid arguments. Use: /sendfile <userId> <filename>")
				continue
			}
			recipientId := args[1]
			filePath := args[2]
			HandleSendFile(conn, recipientId, filePath)
			continue
		case strings.HasPrefix(message, "/status"):
			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println("error in write message", err)
				continue
			}
			continue
		default:
			// Encrypt the message before sending
			encryptedMsg, err := encryption.EncryptMessage(message)
			if err != nil {
				fmt.Println("error encrypting message", err)
				continue
			}
			_, err = conn.Write([]byte(encryptedMsg))
			if err != nil {
				fmt.Println("error in write message", err)
				return
			}
		}
	}
}
