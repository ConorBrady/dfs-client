package dfs

import (
	"github.com/conorbrady/dfs-client/secure"
	"bufio"
	"fmt"
	"regexp"
	"errors"
	"strconv"
	"crypto/sha1"
	"encoding/hex"
	)

const BLOCK_SIZE = 4096

type FileServer struct {
	conn *secure.SecureConn
	reader *bufio.Reader
}

func fsConnect(address string, username string, sessionKey []byte, serverTicket string) (*FileServer, error) {

	conn, err := secure.Connect(address, username, sessionKey, serverTicket)

	if err != nil {
		fmt.Println(err.Error())
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
		return "", nil, errors.New("Error parsing START")
	}

	line, err = fs.reader.ReadString('\n')
	if err != nil {
		return "", nil, err
	}

	r, _ = regexp.Compile("\\AHASH:\\s*(\\S+)\\s*\\z")
	matches = r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", nil, errors.New("Error parsing hash")
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

	data = make([]byte,0)
	buffer := make([]byte,contentLength)
	for len(data) < contentLength {
		n, _ := fs.reader.Read(buffer)
		data = append(data,buffer[ :n]...)
	}

	data = data[ :contentLength]
	hasher := sha1.New()
	hasher.Write(data)
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	if hash != expectedHash {
		return hash, data, nil
		return "", nil, errors.New("Download failed with mismatched hashes")
	}

	return hash, data, nil
}

func (fs * FileServer)write(filename string, start int, data []byte) (string, error) {

	if ( start % BLOCK_SIZE ) + len(data) > BLOCK_SIZE {
		return "", errors.New("Cannot perform writes across blocks")
	}

	if _, err := fs.conn.Write([]byte("WRITE_FILE: "+filename+"\n")); err != nil {
		return "", err
	}

	if _, err := fs.conn.Write([]byte(fmt.Sprintf("START: %d\n",start))); err != nil {
		return "", err
	}

	if _, err := fs.conn.Write([]byte(fmt.Sprintf("CONTENT_LENGTH: %d\n",len(data)))); err != nil {
		return "", err
	}

	if _, err := fs.conn.Write([]byte("DATA:")); err != nil {
		return "", err
	}

	if _, err := fs.conn.Write(data); err != nil {
		return "", err
	}

	if _, err := fs.conn.Write([]byte{'\n'}); err != nil {
		return "", err
	}

	hasher := sha1.New()
	hasher.Write(data)
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	line, _ := fs.reader.ReadString('\n')
	r, _ := regexp.Compile("\\A\000*WROTE_BLOCK:\\s*(\\d+)\\s*\\z")
	if !r.MatchString(line) {
		return "", errors.New("Server did not respond with success")
	}

	line, err := fs.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	r, _ = regexp.Compile("\\A\000*HASH:\\s*(\\S+)\\s*\\z")
	matches := r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", errors.New("Server responded with error")
	}

	hash := matches[1]

	if hash != expectedHash {
		return "", errors.New("Hashes do not match, upload failed")
	}

	fmt.Println("Chunk uploaded")
	return hash, nil
}

func (fs * FileServer)close() {

	fs.conn.Close()
}
