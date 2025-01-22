package connection

import (
	helper "drizlink"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

func HandleSendFolder(conn net.Conn, recipientId, folderPath string) {
	//Create a temperoroy zip file
	tempZipPath := folderPath+".zip"
	err := helper.CreateZipFromFolder(folderPath, tempZipPath)
	if err != nil {
		fmt.Printf("Error creating zip file: ", err)
	}
	defer os.Remove(tempZipPath)  //clean up temporary zip file

	//open zip file
	zipFile, err := os.Open(tempZipPath)
	if err != nil {
		fmt.Printf("Error opening temp zip file: ", err)
		return
	}

	defer zipFile.Close()

	//Get zip file info
    zipInfo ,err := zipFile.Stat()
	if err != nil {
		fmt.Printf("Error getting zip file info: ", err)
		return
	}

	zipSize := zipInfo.Size()
	folderName := filepath.Base(folderPath)

	fmt.Printf("Sending folder '%s' to user %s...\n", folderName, recipientId)
	// Send folder request with zip size
	_, err = conn.Write([]byte(fmt.Sprintf("/FOLDER_REQUEST %s %s %d\n", recipientId, folderName, zipSize)))
	if err != nil {
		fmt.Printf("Error sending folder request: %v\n", err)
		return
	}

	//stream zip file data
	n, err := io.CopyN(conn, zipFile, zipSize)
	if err != nil {
		fmt.Printf("Error sending folder data: %v\n", err)
		return
	}
	if n != zipSize {
		fmt.Printf("Error: sent %d bytes, expected %d bytes\n", n, zipSize)
		return
	}
	fmt.Printf("Folder '%s' sent successfully!\n", folderName)
}
