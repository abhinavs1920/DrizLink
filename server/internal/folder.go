package connection

import (
	"drizlink/server/interfaces"
	"fmt"
	"io"
	"net"
)

func HandleFolderTransfer(server *interfaces.Server, conn net.Conn, recipientId, folderName string, folderSize int64) {
	recipient, exists := server.Connections[recipientId]
	if exists {
		// Send folder transfer response to recipient
		_, err := recipient.Conn.Write([]byte(fmt.Sprintf("/FOLDER_RESPONSE %s %s %d %s", recipientId, folderName, folderSize, recipient.StoreFilePath)))
		if err != nil {
			fmt.Printf("Error sending folder response to %s: %v\n", recipientId, err)
			return
		}

		// Forward the zipped folder data from sender to recipient
		n, err := io.CopyN(recipient.Conn, conn, folderSize)
		if err != nil {
			fmt.Printf("Error transferring folder data: %v\n", err)
			return
		}
		fmt.Printf("Transferred %d bytes of folder data\n", n)
	} else {
		fmt.Printf("User %s not found\n", recipientId)
	}
}

func HandleLookupRequest(server *interfaces.Server, conn net.Conn, userId string) {
	recipient, exists := server.Connections[userId]
	if exists {
		if !recipient.IsOnline {
			fmt.Printf("User %s is not online\n", userId)
			_, err := conn.Write([]byte(fmt.Sprintf("User is not online %s\n", userId)))
			if err != nil {
				fmt.Printf("Error sending lookup response: %v\n", err)
				return
			}
			return
		}
	} else {
		fmt.Printf("User %s not found\n", userId)
		_, err := conn.Write([]byte(fmt.Sprintf("User %s not found\n", userId)))
		if err != nil {
			fmt.Printf("Error sending lookup response: %v\n", err)
			return
		}
	}
}
