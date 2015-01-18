## Overview

This repo contains a client implementation of [Distributed File System](http://github.com/conorbrady/distributed-file-system). It
provides a library written in Go with a simple implementation in app.go for means
of an example.

## Setup Instructions
1. Setup Go
	1. [Install Go](https://golang.org/doc/install) >=1.3.3
	2. Navigate to a directory that will contain your Go directory
	3. Create the folder hierarchy

			mkdir -p go/{src,bin,pkg}
	4. Set this directory to be your GOPATH environment variable from 3.

			export GOPATH="Your/go/directory"
	5. Export go/bin to your bath PATH

			export PATH=$PATH:$GOPATH/bin

4. Download the code

		go get github.com/conorbrady/dfs-client


## Using the library

### Importing
```go
import "github.com/conorbrady/dfs-client/dfs"
```
### Connect to the file system
```go
func Connect(
	authServAdd string, // Address for authentication server
	dirServAdd string, // Address for directory server
	username string, // Username for authentication with the AS
	password string, // Password for authentication with the AS
	caching bool // Enable file caching
) *DFS // Returns an instance that can be used to open files on the system
```
### Open a file for read/write
```go
func (d* DFS)Open(
	filename string // Filename of the file
) (*File, error) // Returns a file or nil and an error

```
### Read a file
```go
// Implements io.Reader Interface
func (f* File)Read(
	p []byte
) (n int, err error)
```
### Write a file
```go
// Implements io.Writer Interface
func (f* File)Write(
	p []byte
) (n int, err error)
```
## Running the Sample

Install the code.

	go install github.com/conorbrady/dfs-client

Once you have added go/bin to your PATH you can run the code from any directory.
Command line options are:

### Authentication Server Address

	-AS <address>

This must be specified to be the address of an instance of the accompanying
[Distributed File System](http://github.com/conorbrady/distributed-file-system) run in `-mode AS`

### Directory Server Address

	-DS <address>

This must be specified to be the address of an instance of the accompanying
[Distributed File System](http://github.com/conorbrady/distributed-file-system) run in `-mode DS`

### Caching Mode

	-caching <0 or 1>

This either enables or disables caching on the system

[Distributed File System]:(http://github.com/conorbrady/distributed-file-system)
[Install Go]:(https://golang.org/doc/install)
