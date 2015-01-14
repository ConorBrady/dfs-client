package main

import(
	"flag"
	"log"
	"os"
	"fmt"
	"bufio"
	"strings"

	"github.com/conorbrady/dfs-client/dfs"
	)

func main (){

	asAddress := flag.String("AS","","Authentication Server Address")
	dsAddress := flag.String("DS","","File Server Address")

	flag.Parse()

	if *asAddress == "" {
		log.Fatal("Authentication Server must be specified")
	}

	if *dsAddress == "" {
		log.Fatal("Directory Server must be specified")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	dfs := dfs.Connect(*asAddress,*dsAddress,username,password)

	file, err := dfs.Create("new2.txt");

	if err != nil {
		log.Fatal(err.Error())
	}

	file.Write([]byte("test file boom with more words\nand a new line\n"))

	if err := file.Close(); err != nil {
		log.Fatal(err.Error())
	}

	differentFile, diffErr := dfs.Open("new2.txt")

	if diffErr != nil {
		log.Fatal(diffErr.Error())
	}

	fileRead := bufio.NewReader(differentFile)
	line, _ := fileRead.ReadString('\n')
	fmt.Println(line)

	line, _ = fileRead.ReadString('\n')
	fmt.Println(line)
}
