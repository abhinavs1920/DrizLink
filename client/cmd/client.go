package main

import (
	connection "drizlink/client/internal"
	"fmt"
)

func main() {
	conn, err := connection.Connect("localhost:8080")
	if err != nil {
		fmt.Println("error in dial")
		panic(err)
	}

	defer connection.Close(conn)

	connection.UserInput("Username", conn)
	connection.UserInput("Store File Path", conn)

	go connection.ReadLoop(conn)

	fmt.Println("\nWelcome to the P2P File Sharing App!")
	fmt.Println("-----------------------------------")
	fmt.Println("You can start typing your messages.")
	fmt.Println("Use '/sendfile <userId> <filename>' to send a file.")
	fmt.Println("Type 'exit' to quit.")

	connection.WriteLoop(conn)
}
