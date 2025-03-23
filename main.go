package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {
	// SFTP server details
	// const (
	// 	host     = "192.168.51.102"
	// 	port     = 822
	// 	user     = "ext-vtt-soc"
	// 	destPath = "/drives/c/Users/ext-vtt-soc/Documents"
	// )

	const (
		host        = "10.30.12.39"
		port        = 822
		user        = "ext-vtt-soc"
		destPath    = "/ThreatHunting"
		passwordHex = "562b684d4f5a63236a35"
	)
	// const (
	// 	host        = "26.173.191.206"
	// 	port        = 22
	// 	user        = "tnt2402"
	// 	destPath    = "/cygdrive/c/Users/tnt2402/"
	// 	passwordHex = "746f6f72"
	// )

	// const (
	// 	host        = "100.100.104.57"
	// 	port        = 822
	// 	user        = "tnt2402"
	// 	destPath    = "/drives/c/Users/tnt2402/AppData/Roaming/MobaXterm/home"
	// 	passwordHex = "746f6f72"
	// )

	// Decode the hex-encoded password
	passwordBytes, err := hex.DecodeString(passwordHex)
	if err != nil {
		fmt.Printf("Failed to decode password: %v\n", err)
		os.Exit(1)
	}
	password := string(passwordBytes)

	// Get the file path from the command-line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: putfile <file_path>")
		os.Exit(1)
	}
	filePath := os.Args[1]

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File does not exist: %s\n", filePath)
		os.Exit(1)
	}

	// Connect to SFTP server
	address := fmt.Sprintf("%s:%d", host, port)
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshConn, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		fmt.Printf("Failed to connect to SFTP server: %v\n", err)
		os.Exit(1)
	}
	defer sshConn.Close()

	sftpClient, err := sftp.NewClient(sshConn)
	if err != nil {
		fmt.Printf("Failed to create SFTP client: %v\n", err)
		os.Exit(1)
	}
	defer sftpClient.Close()

	// Create destination directory if it doesn't exist
	err = sftpClient.MkdirAll(destPath)
	if err != nil {
		fmt.Printf("Failed to create destination directory: %v\n", err)
		os.Exit(1)
	}

	// Open the local file
	localFile, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open local file: %v\n", err)
		os.Exit(1)
	}
	defer localFile.Close()

	// Get the base name of the file to preserve the file name
	fileName := filepath.Base(filePath)
	destFullPath := filepath.Join(destPath, fileName)
	fmt.Println(destFullPath)

	// Create the destination file on the SFTP server
	destFile, err := sftpClient.Create(fileName)
	if err != nil {
		fmt.Printf("Failed to create file on SFTP server: %v\n", err)
		os.Exit(1)
	}
	defer destFile.Close()

	// Copy the file content using io.Copy instead of ReadFrom
	// buffer := make([]byte, 32*1024) // 32KB buffer
	// _, err = io.CopyBuffer(destFile, localFile, buffer)
	// if err != nil {
	// 	fmt.Printf("Failed to upload file: %v\n", err)
	// 	os.Exit(1)
	// }

	_, err = io.Copy(destFile, localFile)
	if err != nil {
		fmt.Printf("Failed to upload file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File successfully uploaded to %s\n", destFullPath)
}
