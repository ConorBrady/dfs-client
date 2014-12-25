package main

import(
	"flag"
	"log"
	"os"
	"fmt"
	"bufio"
	"strings"

	"dfs-client/dfs"
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

	file, err := dfs.Create("~/Desktop/something-test.txt");

	if err != nil {
		log.Fatal(err.Error())
	}

	file.Write([]byte("test file boom\n"))

	if err := file.Close(); err != nil {
		log.Fatal(err.Error())
	}

	differentFile, diffErr := dfs.Open("~/Desktop/something-test.txt")

	if diffErr != nil {
		log.Fatal(diffErr.Error())
	}

	fileRead := bufio.NewReader(differentFile)
	line, _ := fileRead.ReadString('\n')
	fmt.Println(line)
}
