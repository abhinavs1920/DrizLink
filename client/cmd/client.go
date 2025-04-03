package main

import (
	"bufio"
	connection "drizlink/client/internal"
	"drizlink/utils"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func checkServerAvailability(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func promptForServerAddress() string {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Println(utils.InfoColor("Enter server address (format host:port):"))
		fmt.Print(utils.CommandColor(">>> "))
		address, _ := reader.ReadString('\n')
		address = strings.TrimSpace(address)
		
		if !strings.Contains(address, ":") {
			fmt.Println(utils.ErrorColor("❌ Invalid address format. Please use host:port (e.g., localhost:8080)"))
			continue
		}
		
		// Check if server is available at this address
		if !checkServerAvailability(address) {
			fmt.Println(utils.ErrorColor("❌ No server available at " + address))
			fmt.Println(utils.InfoColor("Would you like to try another address? (y/n)"))
			fmt.Print(utils.CommandColor(">>> "))
			
			retry, _ := reader.ReadString('\n')
			retry = strings.TrimSpace(strings.ToLower(retry))
			
			if retry != "y" && retry != "yes" {
				os.Exit(1)
			}
			continue
		}
		
		return address
	}
}

func main() {
	serverAddr := flag.String("server", "", "Server address in format host:port")
	flag.Parse()
	
	utils.PrintBanner()
	
	// If server address not provided via command line, ask user
	address := *serverAddr
	if address == "" {
		address = promptForServerAddress()
	} else {
		fmt.Println(utils.InfoColor("Connecting to server at " + address + "..."))
		
		// Check if server is available
		if !checkServerAvailability(address) {
			fmt.Println(utils.ErrorColor("❌ Error: No server running at " + address))
			fmt.Println(utils.InfoColor("Please check the address or start a server first."))
			return
		}
	}
	
	conn, err := connection.Connect(address)
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
