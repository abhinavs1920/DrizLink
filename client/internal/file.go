package connection

import (
	"drizlink/helper"
	"drizlink/utils"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func HandleSendFile(conn net.Conn, recipientId, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error opening file:"), err)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error getting file info:"), err)
		return
	}

	fileSize := fileInfo.Size()
	fileName := fileInfo.Name()

	// Calculate checksum of file
	checksum, err := helper.CalculateFileChecksum(filePath)
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error calculating checksum:"), err)
		return
	}

	fmt.Printf("%s Sending file '%s' to user %s...\n", 
		utils.InfoColor("📤"),
		utils.InfoColor(fileName), 
		utils.UserColor(recipientId))
	
	// Send file request with file size and checksum
	_, err = conn.Write([]byte(fmt.Sprintf("/FILE_REQUEST %s %s %d %s\n", 
		recipientId, fileName, fileSize, checksum)))
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error sending file request:"), err)
		return
	}

	// Create progress bar
	bar := utils.CreateProgressBar(fileSize, "📤 Sending file")
	
	// Stream file data using io.TeeReader to update progress bar
	n, err := io.CopyN(conn, io.TeeReader(file, bar), fileSize)
	if err != nil {
		fmt.Println(utils.ErrorColor("\n❌ Error sending file:"), err)
		return
	}
	
	if n != fileSize {
		fmt.Println(utils.ErrorColor("\n❌ Error: sent"), utils.ErrorColor(n), 
		    utils.ErrorColor("bytes, expected"), utils.ErrorColor(fileSize), utils.ErrorColor("bytes"))
		return
	}
	
	fmt.Printf("%s File '%s' sent successfully!\n", 
	    utils.SuccessColor("\n✅"),
	    utils.SuccessColor(fileName))
	fmt.Println(utils.InfoColor("  MD5 Checksum:"), utils.InfoColor(checksum))
}

func HandleFileTransfer(conn net.Conn, recipientId, fileName string, fileSize int64, storeFilePath string) {
	// Get checksum from the split content
	parts := strings.SplitN(fileName, "|", 2)
	checksum := ""
	
	if len(parts) == 2 {
		fileName = parts[0]
		checksum = parts[1]
		fmt.Println(utils.InfoColor("📋 Original checksum:"), utils.InfoColor(checksum))
	}
	
	fmt.Printf("%s Receiving file: %s (Size: %s)\n", 
		utils.InfoColor("📥"),
		utils.InfoColor(fileName),
		utils.InfoColor(fmt.Sprintf("%d bytes", fileSize)))

	filePath := filepath.Join(storeFilePath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error creating file:"), err)
		return
	}
	defer file.Close()

	// Create progress bar
	bar := utils.CreateProgressBar(fileSize, "📥 Receiving file")
	
	// Write to file and update progress bar simultaneously
	n, err := io.CopyN(file, io.TeeReader(conn, bar), fileSize)
	if err != nil {
		fmt.Println(utils.ErrorColor("\n❌ Error receiving file:"), err)
		return
	}
	
	if n != fileSize {
		fmt.Println(utils.ErrorColor("\n❌ Error: received"), utils.ErrorColor(n), 
		    utils.ErrorColor("bytes, expected"), utils.ErrorColor(fileSize), utils.ErrorColor("bytes"))
		return
	}
	
	// Verify checksum if provided
	if checksum != "" {
		file.Close() // Close file before calculating checksum
		receivedChecksum, err := helper.CalculateFileChecksum(filePath)
		if err != nil {
			fmt.Println(utils.ErrorColor("\n❌ Error calculating checksum:"), err)
		} else {
			fmt.Println(utils.InfoColor("\n📋 Calculated checksum:"), utils.InfoColor(receivedChecksum))
			
			if helper.VerifyChecksum(checksum, receivedChecksum) {
				fmt.Println(utils.SuccessColor("✅ Checksum verification successful! File integrity confirmed."))
			} else {
				fmt.Println(utils.ErrorColor("❌ Checksum verification failed! File may be corrupted."))
			}
		}
	}

	fmt.Printf("%s File '%s' received successfully!\n", 
	    utils.SuccessColor("✅"),
	    utils.SuccessColor(fileName))
	fmt.Println(utils.InfoColor("📂 Saved to:"), utils.InfoColor(filePath))
}

func HandleDownloadRequest(conn net.Conn, recipientId, filePath string) {
	_, err := conn.Write([]byte(fmt.Sprintf("/DOWNLOAD_REQUEST %s %s\n", recipientId, filePath)))
	if err != nil {
		fmt.Println("Error sending file request:", err)
		return
	}
	fmt.Println("File download request sent successfully")
}

func HandleDownloadResponse(conn net.Conn, userId, filePath string) {
	cleanPath := filepath.Clean(strings.TrimSpace(filePath))
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		fmt.Printf("Error resolving absolute path: %v\n", err)
		return
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		fmt.Println("error in stat file", err)
		return
	}
	if !fileInfo.IsDir() {
		HandleSendFile(conn, userId, absPath)
	} else {
		HandleSendFolder(conn, userId, absPath)
	}
}
