package connection

import (
	"fmt"
	"io"
	"net"
	"os"
)

func HandleSendFile(conn net.Conn, recipientId, filePath string) {
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

	fmt.Printf("Sending file '%s' to user %s...\n", fileName, recipientId)
	// Send file request with file size
	_, err = conn.Write([]byte(fmt.Sprintf("/FILE_REQUEST %s %s %d\n", recipientId, fileName, fileSize)))
	if err != nil {
		fmt.Println("Error sending file request:", err)
		return
	}

	// Stream file data in chunks using io.CopyN
	n, err := io.CopyN(conn, file, fileSize)
	if err != nil {
		fmt.Println("error in write fileData", err)
		return
	}
	if n != fileSize {
		fmt.Println("error in copy file", err)
		return
	}
	fmt.Printf("File '%s' sent successfully with file size %d!\n", fileName, fileSize)
}

func HandleFileTransfer(conn net.Conn, recipientId, fileName string, fileSize int64, storeFilePath string) {
	filePath := storeFilePath + string(os.PathSeparator) + fileName
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("error in create file", err)
		return
	}
	defer file.Close()

	_, err = io.CopyN(file, conn, fileSize)
	if err != nil {
		fmt.Println("error in write fileData", err)
		return
	}
	fmt.Printf("Received file transfer response for: %s (Size: %d bytes)\n", fileName, fileSize)
	fmt.Println("Starting file download...")
	fmt.Printf("File %s successfully received and saved to %s\n", fileName, storeFilePath)
}
