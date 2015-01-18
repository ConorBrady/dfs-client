package dfs

import (
	"github.com/conorbrady/dfs-client/auth"
	"github.com/conorbrady/dfs-client/directory"
	"errors"
	)

type DFS struct {

	sessionKey []byte
	username string
	serverTicket string
	dirServAdd string
	caching bool
}

func Connect(authServAdd string, dirServAdd string, username string, password string, caching bool) *DFS {

	authServ := auth.ConnectToAuthServ(authServAdd)

	sessionKey, serverTicket := authServ.Authenticate(username,password)

	authServ.Close()

	return &DFS{
		sessionKey,
		username,
		serverTicket,
		dirServAdd,
		caching,
	}
}

func (d* DFS)Open(filename string) (*File, error) {

	dir, err := directory.Connect(d.dirServAdd,d.username,d.sessionKey,d.serverTicket)
	if err != nil {
		return nil, err
	}

	fsAddr := dir.Locate(filename)
	dir.Close()


	if fsAddr == nil {
		return nil, errors.New("Could not locate address for fileserver")
	} else {
		return makeFile(*fsAddr, d.username, d.sessionKey, d.serverTicket, filename, d.caching)
	}
}
