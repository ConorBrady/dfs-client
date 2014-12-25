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
}

func Connect(authServAdd string, dirServAdd string, username string, password string) *DFS {

	authServ := auth.ConnectToAuthServ(authServAdd)

	sessionKey, serverTicket := authServ.Authenticate(username,password)

	authServ.Close()

	return &DFS{
		sessionKey,
		username,
		serverTicket,
		dirServAdd,
	}
}

func (d* DFS)Open(filename string) (*File, error) {

	dir, err := directory.Connect(d.dirServAdd,d.username,d.sessionKey,d.serverTicket)
	if err != nil {
		dir.Close()
		return nil, err
	}

	fsAddr := dir.Locate(filename)
	dir.Close()

	if fsAddr == nil {
		return nil, errors.New("Could not locate address for fileserver")
	} else {
		return MakeFile(*fsAddr, d.username, d.sessionKey, d.serverTicket, filename, false)
	}
}

func (d* DFS)Create(filename string) (*File, error) {

	dir, err := directory.Connect(d.dirServAdd,d.username,d.sessionKey,d.serverTicket)
	if err != nil {
		dir.Close()
		return nil, err
	}

	fsAddr := dir.Locate(filename)
	dir.Close()


	if fsAddr == nil {
		return nil, errors.New("Could not locate address for fileserver")
	} else {
		return MakeFile(*fsAddr, d.username, d.sessionKey, d.serverTicket, filename, true)
	}
}
