package dfs

import (
	"time"
	"fmt"
	"io"
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
	seekHead int
	cache []CacheBlock
	caching bool
}

func MakeFile(fsAddr string, username string, sessionKey []byte, serverTicket string, name string, caching bool) ( *File, error ) {

	f := &File{
		fsAddr,
		username,
		sessionKey,
		serverTicket,
		name,
		0,
		make([]CacheBlock,0),
		caching,
	}

	if caching {
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for {
				<- ticker.C
				fmt.Println("Validating cache")
				f.fs().validateCache(f.name,f.cache)
			}
		}()
	}

	return f, nil
}

func (f* File)fs() *FileServer{
	fs, _ := fsConnect(f.fsAddr,f.username,f.sessionKey,f.serverTicket)
	return fs
}

func (f* File)Read(p []byte) (n int, err error) {

	var data []byte = nil
	blockIndex := f.seekHead/BLOCK_SIZE

	if len(f.cache) > blockIndex && f.cache[blockIndex].valid == true && f.caching {

		data = f.cache[blockIndex].data

	} else {

		fmt.Printf("Requesting %d\n",blockIndex)
		hash, read, err := f.fs().read(f.name,blockIndex)
		data = read
		if err != nil {
			fmt.Println("A read error occured: "+err.Error())
			return 0, err
		}

		if f.caching { // Add to cache

			for len(f.cache) <= blockIndex {
				f.cache = append(f.cache,CacheBlock{"",false,nil})
			}

			f.cache[blockIndex].hash = hash
			f.cache[blockIndex].data = data
			f.cache[blockIndex].valid = true
		}
	}
	fmt.Printf("Got data %d bytes",len(data))
	n = copy(p,data[f.seekHead%BLOCK_SIZE:])
	f.seekHead = f.seekHead + n

	if n == 0 {
		return 0, io.EOF
	} else {
		return n, nil
	}
}

func (f* File)Write(data []byte) (n int, err error) {

	n = len(data)

	for len(data) > 0 {

		cut := BLOCK_SIZE - f.seekHead % BLOCK_SIZE

		if len(data) < cut {
			cut = len(data)
		}

		chunk := data[ :cut]
		data = data[cut: ]

		hash, err := f.fs().write(f.name, f.seekHead, chunk)

		if err != nil{
			return 0, err
		}

		if f.caching { // Update cache

			blockIndex := f.seekHead/BLOCK_SIZE

			if blockIndex < len(f.cache) && f.cache[blockIndex].valid {
				f.cache[blockIndex].hash = hash
				copy(f.cache[blockIndex].data[f.seekHead%BLOCK_SIZE:],chunk)
			}
		}

		f.seekHead = f.seekHead + cut
	}

	return n, nil
}

func (f* File)Close() error {

	return nil
}
