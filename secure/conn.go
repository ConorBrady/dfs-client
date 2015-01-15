package secure

import (
	"bufio"
	"net"
	"time"
	"strconv"
	"github.com/conorbrady/dfs-client/crypto"
	"encoding/base64"
	"regexp"
	"errors"
	)

type SecureConn struct {
	encryptedConnReader *bufio.Reader
	encryptedConn *net.TCPConn
	sessionKey []byte
	readBuffer chan byte
	writeBuffer chan byte
}

func Connect(address string, username string, sessionKey []byte, serverTicket string) (*SecureConn, error) {

	secureServAddr, ssAddrErr := net.ResolveTCPAddr("tcp4",address)

	if ssAddrErr != nil {
		return nil, errors.New("Valid Secure Server address must be specified, recieved "+address)
	}

	encryptedConn, ssConnErr := net.DialTCP("tcp4",nil,secureServAddr)

	if ssConnErr != nil {
		return nil, errors.New("Could not connect to secure server")
	}

	encryptedConnReader := bufio.NewReader(encryptedConn)

	sentTimestamp := time.Now().Unix()

	authenticator := "USERNAME: " + username + "\n" +
					 "TIMESTAMP: " + strconv.Itoa(int(sentTimestamp)) + "\n"

	encAuthenticator := crypto.EncryptString(authenticator,sessionKey)

	encryptedConn.Write([]byte("SERVICE_TICKET: "+serverTicket+"\n"))
	encryptedConn.Write([]byte("AUTHENTICATOR: "+base64.StdEncoding.EncodeToString(encAuthenticator)+"\n"))


	line, _ := encryptedConnReader.ReadString('\n')

	rgx, _ := regexp.Compile("\\ARESPONSE:\\s*(\\S+)\\s*\\z")
	matches := rgx.FindStringSubmatch(line)

	if len(matches) < 2 {
		return nil, errors.New("Response not recieved")
	}

	encResponse, encResErr := base64.StdEncoding.DecodeString(matches[1])

	if encResErr != nil {
		return nil, errors.New("Response not valid base64")
	}

	incrementedTimestamp, incTimeErr := strconv.Atoi(crypto.DecryptToString(encResponse,sessionKey))

	if incTimeErr != nil {

		return nil, errors.New("Response timestamp not an int")
	}

	if  incrementedTimestamp - 1 != int(sentTimestamp) {
		return nil, errors.New("Server cannot be trusted")
	}

	ss := &SecureConn{
		encryptedConnReader,
		encryptedConn,
		sessionKey,
		make(chan byte, 64),
		make(chan byte, 64),
	}

	go ss.ReadLoop()
	go ss.WriteLoop()

	return ss, nil
}

func (ss* SecureConn)ReadLoop() {

	for {

		line, _ := ss.encryptedConnReader.ReadString('\n')
		enc, _ := base64.StdEncoding.DecodeString(line)
		for _, b := range crypto.DecryptToBytes(enc,ss.sessionKey) {
			ss.readBuffer <- b
		}
	}
}

func (ss* SecureConn)WriteLoop() {

	for {

		data := make([]byte,32)
		for i, _ := range data {
			b := <-ss.writeBuffer
			data[i] = b

			if b == '\n' {
				break
			}
		}
		enc := crypto.EncryptBytes(data,ss.sessionKey)

		ss.encryptedConn.Write([]byte(base64.StdEncoding.EncodeToString(enc)+"\n"))
	}
}

func (ss* SecureConn)Read(p []byte) (n int, err error) {

	n = 1
	err = nil
	b := <-ss.readBuffer
	p[0] = b

	for i := 1; i < len(p); i += 1 {
		select {
			case b = <-ss.readBuffer:
				p[i] = b
				n += 1
			default:
				if n != 0 {
					return
				}
		}
	}
	return
}

func (ss* SecureConn)Write(p []byte) (n int, err error) {

	for _, b := range p {
		ss.writeBuffer <- b
	}
	return len(p), nil
}

func (ss* SecureConn)Close() {
	ss.encryptedConn.Close()
}
