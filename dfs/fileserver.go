package dfs

import (
	"github.com/conorbrady/dfs-client/secure"
	"bufio"
	"fmt"
	"regexp"
	"errors"
	"strconv"
	)

const BLOCK_SIZE = 4096

type FileServer struct {
	conn *secure.SecureConn
	reader *bufio.Reader
}

func fsConnect(address string, username string, sessionKey []byte, serverTicket string) (*FileServer, error) {

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

func (fs * FileServer)validateCache(filename string, cache []CacheBlock) error {

	fs.conn.Write([]byte("REQUEST_CHECKSUMS: "+filename))

	for i, cb := range cache {
		if cb.valid {
			fs.conn.Write([]byte("INDEX: "+strconv.Itoa(i)+"\n"))
		}
	}
	fs.conn.Write([]byte("END_REQUEST_CHECKSUMS"))

	line, err := fs.reader.ReadString('\n')
	if err != nil {
		return err
	}

	r, _ := regexp.Compile("\\ACHECKSUMS:\\s*(\\S+)\\s*\\z")
	matches := r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return errors.New("Server responded with error")
	}

	for {

		line, err = fs.reader.ReadString('\n')
		if err != nil {
			return err
		}

		r, _ = regexp.Compile("\\AINDEX:\\s*(\\d+)\\s*\\z")
		matches = r.FindStringSubmatch(line)
		if len(matches) < 2 {
			break // end of cache validation, assume for this, would need more robust if for reals
		}

		index, _ := strconv.Atoi(matches[1])

		line, err = fs.reader.ReadString('\n')
		if err != nil {
			return err
		}

		r, _ = regexp.Compile("\\AHASH:\\s*(\\S+)\\s*\\z")
		matches = r.FindStringSubmatch(line)
		if len(matches) < 2 {
			return errors.New("Server responded with error")
		}

		hash := matches[1]

		if cache[index].hash != hash {
			cache[index].valid = false
		}
	}
	return nil
}

func (fs * FileServer)read(filename string, block_index int) (hash string, data []byte, err error) {

	fs.conn.Write([]byte("READ_FILE: "+filename+"\n"))
	fs.conn.Write([]byte("OFFSET: "+strconv.Itoa(block_index*BLOCK_SIZE)+"\n"))

	line, err := fs.reader.ReadString('\n')
	if err != nil {
		return "", nil, err
	}

	r, _ := regexp.Compile("\\ASTART:\\s*(\\d+)\\s*\\z")
	matches := r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", nil, errors.New("Server responded with error")
	}

	line, err = fs.reader.ReadString('\n')
	if err != nil {
		return "", nil, err
	}

	r, _ = regexp.Compile("\\AHASH:\\s*(\\S+)\\s*\\z")
	matches = r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", nil, errors.New("Server responded with error")
	}

	hash = matches[1]

	line, err = fs.reader.ReadString('\n')
	if err != nil {
		return "", nil, err
	}

	r, _ = regexp.Compile("\\ACONTENT_LENGTH\\s*:\\s*(\\d+)\\s*\\z")
	matches = r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", nil, errors.New("Server responded with error")
	}

	contentLength, _ := strconv.Atoi(matches[1])

	buffer := make([]byte,contentLength)

	_, bErr := fs.reader.Read(buffer)

	return hash, buffer, bErr
}

func (fs * FileServer)write(filename string, start int, data []byte) error {

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

func (fs * FileServer)close() {

	fs.conn.Close()
}
