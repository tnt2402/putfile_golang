package main

import (
	"encoding/hex"
	"fmt"
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
		host     = "10.30.12.39"
		port     = 822
		user     = "ext-vtt-soc"
		destPath = "/ThreatHunting"
	)

	// Hex-encoded password
	passwordHex := "562b684d4f5a63236a35"

	// Decode the hex-encoded password
	passwordBytes, err := hex.DecodeString(passwordHex)
	if err != nil {
		fmt.Printf("Failed to decode password: %v\n", err)
		os.Exit(1)
	}
	password := string(passwordBytes)

	// Get the file path from the command-line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run script.go <file_path>")
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

	// Create the destination file on the SFTP server
	destFile, err := sftpClient.Create(destFullPath)
	if err != nil {
		fmt.Printf("Failed to create file on SFTP server: %v\n", err)
		os.Exit(1)
	}
	defer destFile.Close()

	// Copy the file content
	_, err = destFile.ReadFrom(localFile)
	if err != nil {
		fmt.Printf("Failed to upload file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File successfully uploaded to %s\n", destFullPath)
}
