package directory

import (
	"github.com/conorbrady/dfs-client/secure"
	"bufio"
	"fmt"
	"regexp"
	)

type Directory struct {
	conn *secure.SecureConn
	reader *bufio.Reader
}

func Connect(address string, username string, sessionKey []byte, serverTicket string) (*Directory, error) {
	conn, err := secure.Connect(address, username, sessionKey, serverTicket)

	if err != nil {
		return nil, err
	}

	return &Directory{
		conn,
		bufio.NewReader(conn),
	}, nil
}

func (d * Directory)Locate(filename string) *string {

	d.conn.Write([]byte("LOCATE: "+filename+"\n"))

	line, err := d.reader.ReadString('\n')
	if err != nil {
		fmt.Print(err.Error())
	}

	r, _ := regexp.Compile("\\AADDRESS:\\s*(\\S+)\\s*\\z")
	matches := r.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil
	}

	return &matches[1]
}

func (d * Directory)Close() {
	d.conn.Close()
}
