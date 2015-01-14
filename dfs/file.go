package dfs

import (
	"errors"
	"time"
	"fmt"
)

type CacheBlock struct {
	hash string
	valid bool
	data []byte
}

type File struct {
	fsAddr string
	username string
	sessionKey []byte
	serverTicket string
	name string
	writeEnable bool
	seekHead int
	cache []CacheBlock
}

func MakeFile(fsAddr string, username string, sessionKey []byte, serverTicket string, name string, writeEnable bool) ( *File, error ) {

	f := &File{
		fsAddr,
		username,
		sessionKey,
		serverTicket,
		name,
		writeEnable,
		0,
		make([]CacheBlock,0),
	}

	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for {
			<- ticker.C
			fmt.Println("Validating cache")
			f.fs().validateCache(f.name,f.cache)
		}
	}()

	return f, nil
}

func (f* File)fs() *FileServer{
	fs, _ := fsConnect(f.fsAddr,f.username,f.sessionKey,f.serverTicket)
	return fs
}

func (f* File)Read(p []byte) (n int, err error) {

	if !f.writeEnable {

		var data []byte = nil
		blockIndex := f.seekHead/BLOCK_SIZE

		if len(f.cache) > blockIndex && f.cache[blockIndex].valid == true {

			data = f.cache[blockIndex].data

		} else {

			hash, data, err := f.fs().read(f.name,f.seekHead/BLOCK_SIZE)
			if err != nil {
				return 0, err
			}

			for len(f.cache) <= blockIndex {
				f.cache = append(f.cache,CacheBlock{"",false,nil})
			}

			f.cache[blockIndex].hash = hash
			f.cache[blockIndex].data = data
			f.cache[blockIndex].valid = true
		}

		n = copy(p,data[f.seekHead%BLOCK_SIZE:])
		f.seekHead = f.seekHead + n

		return n, nil

	} else {

		return 0, errors.New("Cannot read in write mode")
	}
}

func (f* File)Write(p []byte) (n int, err error) {

	if f.writeEnable {

		if err := f.fs().write(f.name,f.seekHead,p); err != nil{
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
