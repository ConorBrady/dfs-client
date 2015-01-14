package dfs

import (
	"errors"
	"github.com/conorbrady/dfs-client/fileserver"
)

type File struct {
	fsAddr string
	username string
	sessionKey []byte
	serverTicket string
	name string
	writeEnable bool
	seekHead int
}

func MakeFile(fsAddr string, username string, sessionKey []byte, serverTicket string, name string, writeEnable bool) ( *File, error ) {

	return &File{
		fsAddr,
		username,
		sessionKey,
		serverTicket,
		name,
		writeEnable,
		0,
	}, nil
}

func (f* File)fs() *fileserver.FileServer{
	fs, _ := fileserver.Connect(f.fsAddr,f.username,f.sessionKey,f.serverTicket)
	return fs
}

func (f* File)Read(p []byte) (n int, err error) {

	if !f.writeEnable {

		start, data, err := f.fs().Read(f.name,f.seekHead)
		if err != nil {
			return 0, err
		}

		n := copy(p,data[f.seekHead-start:])
		f.seekHead = f.seekHead + n
		return n, nil

	} else {
		return 0, errors.New("Cannot read in write mode")
	}
}

func (f* File)Write(p []byte) (n int, err error) {

	if f.writeEnable {

		if err := f.fs().Write(f.name,f.seekHead,p); err != nil{
			return 0, err
		}

		n := len(p)

		f.seekHead = f.seekHead + n
		return n, nil

	} else {
		return 0, errors.New("Cannot write in read mode")
	}
}

func (f* File)Close() error {

	return nil
}
