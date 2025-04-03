package main

import (
	"drizlink/server/interfaces"
	connection "drizlink/server/internal"
	"drizlink/utils"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"
)

// Check if port is already in use
func isPortInUse(port string) bool {
	conn, err := net.DialTimeout("tcp", "localhost:"+port, time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()
	
	// Ensure port starts with a colon for address format
	formattedPort := *port
	if !strings.HasPrefix(formattedPort, ":") {
		formattedPort = ":" + formattedPort
	}
	
	// Check if port is already in use
	if isPortInUse(*port) {
		fmt.Println(utils.ErrorColor("❌ Error: Port " + *port + " is already in use"))
		fmt.Println(utils.InfoColor("Please choose a different port or stop the other server."))
		return
	}
	
	utils.PrintBanner()
	fmt.Println(utils.InfoColor("Starting server on port " + *port + "..."))
	
	server := interfaces.Server{
		Address:     formattedPort,
		Connections: make(map[string]*interfaces.User),
		IpAddresses: make(map[string]*interfaces.User),
		Messages:    make(chan interfaces.Message),
	}

	go connection.StartHeartBeat(100*time.Second, &server)
	connection.Start(&server)
}
