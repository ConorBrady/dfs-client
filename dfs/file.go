package dfs

import (
	"code.google.com/p/go-uuid/uuid"
	"os"
	"github.com/conorbrady/dfs-client/fileserver"
)

type File struct {
	fsAddr string
	username string
	sessionKey []byte
	serverTicket string
	name string
	writeEnable bool
	uuid string
	file *os.File
}

func MakeFile(fsAddr string, username string, sessionKey []byte, serverTicket string, name string, writeEnable bool) ( *File, error ) {

	uuid := uuid.New()

	var file *os.File = nil

	if !writeEnable {

		fs, fsErr := fileserver.Connect(fsAddr,username,sessionKey,serverTicket)

		if fsErr != nil {
			return nil, fsErr
		}

		if err := fs.Pull(name,os.TempDir()+uuid); err != nil {
			return nil, err
		}

		fs.Close()

		file, _ = os.Open(os.TempDir()+uuid)

	} else {

		file, _ = os.Create(os.TempDir()+uuid)
	}

	return &File{
		fsAddr,
		username,
		sessionKey,
		serverTicket,
		name,
		writeEnable,
		uuid,
		file,
	}, nil
}

func (f* File)Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

func (f* File)Write(p []byte) (n int, err error) {
	return f.file.Write(p)
}

func (f* File)Close() error {

	if f.writeEnable {

		fs, fsErr := fileserver.Connect(f.fsAddr,f.username,f.sessionKey,f.serverTicket)

		if fsErr != nil {
			return fsErr
		}

		if err := fs.Push(f.name,os.TempDir()+f.uuid); err != nil {
			return err
		}

		fs.Close()
	}

	return f.file.Close()
}
