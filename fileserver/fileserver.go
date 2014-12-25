package fileserver

import (
	"dfs-client/secure"
	"bufio"
	"fmt"
	"regexp"
	"errors"
	"os"
	"strconv"
	)

type FileServer struct {
	conn *secure.SecureConn
	reader *bufio.Reader
}

func Connect(address string, username string, sessionKey []byte, serverTicket string) (*FileServer, error) {

	conn, err := secure.Connect(address, username, sessionKey, serverTicket)

	if err != nil {
		conn.Close()
		return nil, err
	}

	return &FileServer{
		conn,
		bufio.NewReader(conn),
	}, nil
}

func (fs * FileServer)Pull(filename string, tempFilename string) error {

	fs.conn.Write([]byte("READ_FILE: "+filename+"\n"))

	line, err := fs.reader.ReadString('\n')
	if err != nil {
		return err
	}

	r, _ := regexp.Compile("\\ACONTENT_LENGTH:\\s*(\\d+)\\s*\\z")
	matches := r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return errors.New("Server responded with error")
	}

	contentLength, _ := strconv.Atoi(matches[1])

	tempFile, _ := os.Create(tempFilename)

	buffer := make([]byte,128)
	readCount := 0

	for contentLength > readCount {
		n, _ := fs.reader.Read(buffer)
		readCount += n
		tempFile.Write(buffer[:n])
	}

	return tempFile.Close()
}

func (fs * FileServer)Push(filename string, tempFilename string) error {

	tempFile, err := os.Open(tempFilename)

	if err != nil {
		return err
	}

	fi, _ := tempFile.Stat()
	contentLength := fi.Size()

	if _, err := fs.conn.Write([]byte("WRITE_FILE: "+filename+"\n")); err != nil {
		return err
	}

	if _, err := fs.conn.Write([]byte(fmt.Sprintf("CONTENT_LENGTH: %d\n",fi.Size()))); err != nil {
		return err
	}

	reader := bufio.NewReader(tempFile)

	buffer := make([]byte,128)
	readCount := int64(0)

	for contentLength > readCount {
		n, _ := reader.Read(buffer)
		readCount += int64(n)
		fs.conn.Write(buffer[:n])
	}

	line, _ := fs.reader.ReadString('\n')
	r, _ := regexp.Compile("\\A\\s*SUCCESS\\s*\\z")
	if !r.MatchString(line) {
		return errors.New("Server did not respond with success")
	}

	return tempFile.Close()

}

func (fs * FileServer)Close() {

	fs.conn.Close()
}
