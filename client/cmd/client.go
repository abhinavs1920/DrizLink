package main

import (
	connection "drizlink/client/internal"
	"drizlink/utils"
	"fmt"
)

func main() {
	utils.PrintBanner()
	fmt.Println(utils.InfoColor("Connecting to server..."))
	
	conn, err := connection.Connect(":8080")
	if err != nil {
		if err.Error() == "reconnect" {
			goto startChat
		} else {
			fmt.Println(utils.ErrorColor("❌ Error connecting to server:"), err)
			return
		}
	}

	defer connection.Close(conn)

	fmt.Println(utils.InfoColor("Please login to continue:"))
	err = connection.UserInput("Username", conn)
	if err != nil {
		if err.Error() == "reconnect" {
			goto startChat
		} else {
			fmt.Println(utils.ErrorColor("❌ Error during login:"), err)
			return
		}
	}

	// Only ask for store file path for new connections
	err = connection.UserInput("Store File Path", conn)
	if err != nil {
		if err.Error() == "reconnect" {
			goto startChat
		} else {
			fmt.Println(utils.ErrorColor("❌ Error setting file path:"), err)
			return
		}
	}

startChat:
	fmt.Println(utils.HeaderColor("\n✨ Welcome to DrizLink - P2P File Sharing! ✨"))
	fmt.Println(utils.InfoColor("------------------------------------------------"))
	fmt.Println(utils.SuccessColor("✅ Successfully connected to server!"))
	fmt.Println(utils.InfoColor("Type /help to see available commands"))
	fmt.Println(utils.InfoColor("------------------------------------------------"))

	go connection.ReadLoop(conn)
	connection.WriteLoop(conn)
}
