package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jlaffaye/ftp"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {
	// Check if enough arguments are provided
	if len(os.Args) < 7 {
		fmt.Println("Usage: putfile <mode> <host> <port> <user> <base64_password> <dest_path> <file_path>")
		fmt.Println("Mode: 'sftp' or 'ftp'")
		fmt.Println("Example: putfile sftp 10.30.12.39 822 ext-vtt-soc VjJoTWFXNW= /ThreatHunting ./myfile.txt")
		os.Exit(1)
	}

	// Get variables from command-line arguments
	mode := os.Args[1]
	host := os.Args[2]
	var port int
	_, err := fmt.Sscanf(os.Args[3], "%d", &port)
	if err != nil {
		fmt.Printf("Invalid port number: %v\n", err)
		os.Exit(1)
	}
	user := os.Args[4]
	base64Password := os.Args[5]
	destPath := os.Args[6]
	filePath := os.Args[7]

	// Decode the base64-encoded password
	passwordBytes, err := base64.StdEncoding.DecodeString(base64Password)
	if err != nil {
		fmt.Printf("Failed to decode base64 password: %v\n", err)
		os.Exit(1)
	}
	password := string(passwordBytes)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File does not exist: %s\n", filePath)
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
	destFullPath := destPath + "/" + fileName
	fmt.Println("Dest File: " + destFullPath)

	switch mode {
	case "sftp":
		err = uploadViaSFTP(host, port, user, password, destPath, filePath, destFullPath, localFile)
	case "ftp":
		err = uploadViaFTP(host, port, user, password, destPath, filePath, destFullPath, localFile)
	default:
		fmt.Printf("Invalid mode: %s. Use 'sftp' or 'ftp'\n", mode)
		os.Exit(1)
	}

	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("File successfully uploaded to %s\n", destFullPath)
}

func uploadViaSFTP(host string, port int, user, password, destPath, filePath, destFullPath string, localFile *os.File) error {
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
		return err
	}
	defer sshConn.Close()
	fmt.Println("[+] Connect to SFTP server successfully")

	sftpClient, err := sftp.NewClient(sshConn)
	if err != nil {
		fmt.Printf("Failed to create SFTP client: %v\n", err)
		return err
	}
	defer sftpClient.Close()

	// Create destination directory if it doesn't exist
	err = sftpClient.MkdirAll(destPath)
	if err != nil {
		fmt.Printf("Failed to create destination directory: %v\n", err)
		return err
	}

	// Create the destination file on the SFTP server
	destFile, err := sftpClient.Create(destFullPath)
	if err != nil {
		fmt.Printf("Failed to create file on SFTP server: %v\n", err)
		return err
	}
	defer destFile.Close()

	// Copy the file content
	_, err = io.Copy(destFile, localFile)
	if err != nil {
		fmt.Printf("Failed to upload file: %v\n", err)
		return err
	}

	return nil
}

func uploadViaFTP(host string, port int, user, password, destPath, filePath, destFullPath string, localFile *os.File) error {
	// Connect to FTP server
	address := fmt.Sprintf("%s:%d", host, port)
	ftpConn, err := ftp.Dial(address)
	if err != nil {
		fmt.Printf("Failed to connect to FTP server: %v\n", err)
		return err
	}
	defer ftpConn.Quit()

	// Login to FTP server
	err = ftpConn.Login(user, password)
	if err != nil {
		fmt.Printf("Failed to login to FTP server: %v\n", err)
		return err
	}

	// Change to destination directory (or create it)
	err = ftpConn.ChangeDir(destPath)
	if err != nil {
		// Try to create the directory if it doesn't exist
		err = ftpConn.MakeDir(destPath)
		if err != nil {
			fmt.Printf("Failed to create destination directory: %v\n", err)
			return err
		}
		err = ftpConn.ChangeDir(destPath)
		if err != nil {
			fmt.Printf("Failed to change to destination directory: %v\n", err)
			return err
		}
	}

	fileName := filepath.Join(destFullPath)
	// Upload the file
	err = ftpConn.Stor(fileName, localFile)
	if err != nil {
		fmt.Printf("Failed to upload file via FTP: %v\n", err)
		return err
	}

	return nil
}
