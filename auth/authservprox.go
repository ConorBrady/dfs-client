package auth

import (
	"net"
	"bufio"
	"log"
	"regexp"
	"encoding/base64"
	"crypto/sha256"
	"strings"
	"crypto/aes"
	)

type AuthServProxy struct {
	authServConnReader *bufio.Reader
	authServConn *net.TCPConn
}

func ConnectToAuthServ(address string) *AuthServProxy {

	// Authentication Server Communication

	authServAddr, asAddrErr := net.ResolveTCPAddr("tcp4",address)

	if asAddrErr != nil {
		log.Fatal("Valid Authentication Server address must be specified")
	}

	authServConn, asConnErr := net.DialTCP("tcp4",nil,authServAddr)

	if asConnErr != nil {
		log.Fatal("Could not connect to authentication server")
	}

	authServConnReader := bufio.NewReader(authServConn)

	return &AuthServProxy{
		authServConnReader,
		authServConn,
	}
}

func (as* AuthServProxy)Authenticate(username string, password string) ([]byte, string) {

	as.authServConn.Write([]byte("AUTHENTICATE:"+username+"\n"))

	line, _ := as.authServConnReader.ReadString('\n')
	rgx, _ := regexp.Compile("\\AENCRYPTED_SESSION_KEY:\\s*(\\S+)\\s*\\z")
	matches := rgx.FindStringSubmatch(line)

	if len(matches) < 2 {
		log.Fatal("User not found")
	}

	encryptedKey, encKeyErr := base64.StdEncoding.DecodeString(matches[1])

	if encKeyErr != nil {
		log.Fatal("Key not valid base64")
	}

	passwordHash := sha256.Sum256([]byte(strings.TrimSpace(password)))

	cipherBlock, _ := aes.NewCipher(passwordHash[:])

	sessionKey := make([]byte,32)
	cipherBlock.Decrypt(sessionKey[0:16], encryptedKey[0:16])
	cipherBlock.Decrypt(sessionKey[16:32],encryptedKey[16:32])

	line, _ = as.authServConnReader.ReadString('\n')
	rgx, _ = regexp.Compile("\\ASERVICE_TICKET:\\s*(\\S+)\\s*\\z")
	matches = rgx.FindStringSubmatch(line)

	if len(matches) < 2 {
		log.Fatal("Server ticket not recieved")
	}

	serverTicket := matches[1]

	return sessionKey, serverTicket
}

func (as* AuthServProxy)Close() {
	as.authServConn.Close()
}
