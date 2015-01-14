package fileserver

import (
	"github.com/conorbrady/dfs-client/secure"
	"bufio"
	"fmt"
	"regexp"
	"errors"
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

func (fs * FileServer)Read(filename string, offset int) (start int, data []byte, err error) {

	fs.conn.Write([]byte("READ_FILE: "+filename+"\n"))
	fs.conn.Write([]byte("OFFSET: "+strconv.Itoa(offset)+"\n"))

	line, err := fs.reader.ReadString('\n')
	if err != nil {
		return 0, nil, err
	}

	r, _ := regexp.Compile("\\ASTART:\\s*(\\d+)\\s*\\z")
	matches := r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return 0, nil, errors.New("Server responded with error")
	}

	start, _ = strconv.Atoi(matches[1])

	line, err = fs.reader.ReadString('\n')
	if err != nil {
		return 0, nil, err
	}

	r, _ = regexp.Compile("\\AHASH:\\s*(\\S+)\\s*\\z")
	matches = r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return 0, nil, errors.New("Server responded with error")
	}

	//hash := matches[1]

	line, err = fs.reader.ReadString('\n')
	if err != nil {
		return 0, nil, err
	}

	r, _ = regexp.Compile("\\ACONTENT_LENGTH\\s*:\\s*(\\d+)\\s*\\z")
	matches = r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return 0, nil, errors.New("Server responded with error")
	}

	contentLength, _ := strconv.Atoi(matches[1])

	buffer := make([]byte,contentLength)

	_, bErr := fs.reader.Read(buffer)

	return start, buffer, bErr
}

func (fs * FileServer)Write(filename string, start int, data []byte) error {

	if _, err := fs.conn.Write([]byte("WRITE_FILE: "+filename+"\n")); err != nil {
		return err
	}

	if _, err := fs.conn.Write([]byte(fmt.Sprintf("START: %d\n",start))); err != nil {
		return err
	}

	if _, err := fs.conn.Write([]byte(fmt.Sprintf("CONTENT_LENGTH: %d\n",len(data)))); err != nil {
		return err
	}

	if _, err := fs.conn.Write(data); err != nil {
		return err
	}

	line, _ := fs.reader.ReadString('\n')
	r, _ := regexp.Compile("\\A\\s*WROTE_BLOCK:\\s*(\\S+)\\s*\\z")
	if !r.MatchString(line) {
		return errors.New("Server did not respond with success")
	}

	line, _ = fs.reader.ReadString('\n')
	r, _ = regexp.Compile("\\A\\s*HASH:\\s*(\\S+)\\s*\\z")
	if !r.MatchString(line) {
		return errors.New("Server did not respond with success")
	}
	return nil
}

func (fs * FileServer)Close() {

	fs.conn.Close()
}
