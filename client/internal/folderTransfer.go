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

func HandleFolderTransfer(conn net.Conn, recipientId, folderName string, folderSize int64, storeFilePath string) {
	fmt.Printf("Receiving folder: %s (Size: %d bytes)\n", folderName, folderSize)

	// Create temporary zip file to store received data
	tempZipPath := filepath.Join(storeFilePath, folderName+".zip")
	zipFile, err := os.Create(tempZipPath)
	if err != nil {
		fmt.Printf("Error creating temporary zip file: %v\n", err)
		return
	}

	// Receive the zip file data
	n, err := io.CopyN(zipFile, conn, folderSize)
	if err != nil {
		zipFile.Close()
		os.Remove(tempZipPath)
		fmt.Printf("Error receiving folder data: %v\n", err)
		return
	}
    zipFile.Close()

	if n!= folderSize {
		os.Remove(tempZipPath)
		fmt.Printf("Error: received %d bytes, expected %d bytes\n", n, folderSize)
		return
	}

	//Extract the zip file
	destPath := filepath.Join(storeFilePath, folderName)
	err = helper.ExtractZip(tempZipPath, destPath)
	if err != nil {
		os.Remove(tempZipPath)
		fmt.Printf("Error extracting folder: %v\n", err)
		return
	}

	// Clean up the temporary zip file
	os.Remove(tempZipPath)
	fmt.Printf("Folder '%s' received and extracted successfully!\n", folderName)
}