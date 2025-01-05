package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("error in dial")
		panic(err)
	}

	fmt.Println("Enter your userId: ")
	reader := bufio.NewReader(os.Stdin)
	userId, _ := reader.ReadString('\n')
	userId = userId[:len(userId)-1]

	fmt.Println("Enter your username: ")
	username, _ := reader.ReadString('\n')
	username = username[:len(username)-1]

	_, err = conn.Write([]byte(userId))
	if err != nil {
		fmt.Println("error in write")
		panic(err)
	}

}
