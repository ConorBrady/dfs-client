package main

import(
	"flag"
	"log"
	"os"
	"bufio"

	"github.com/conorbrady/dfs-client/dfs"
	)

func main (){

	asAddress := flag.String("AS","","Authentication Server Address")
	dsAddress := flag.String("DS","","File Server Address")
	caching := flag.Int("caching",1,"Turn on caching")

	flag.Parse()

	if *asAddress == "" {
		log.Fatal("Authentication Server must be specified")
	}

	if *dsAddress == "" {
		log.Fatal("Directory Server must be specified")
	}


	// reader := bufio.NewReader(os.Stdin)

	// fmt.Print("Username: ")
	// username, _ := reader.ReadString('\n')
	// username = strings.TrimSpace(username)
	//
	// fmt.Print("Password: ")
	// password, _ := reader.ReadString('\n')
	// password = strings.TrimSpace(password)

	dfs := dfs.Connect(*asAddress,*dsAddress,"conorbrady","password",*caching>0)

	fileRemote, rErr := dfs.Open("planet2.jpg")

	if rErr != nil {
		log.Fatal(rErr.Error())
	}

	fileLocal, lErr := os.Create("planet4.jpg")

	if lErr != nil {
		log.Fatal(lErr.Error())
	}

	bufio.NewReader(fileRemote).WriteTo(fileLocal)

	// file.Write([]byte("test file boom with more words\nand a new line\n"))
	//
	// if err := file.Close(); err != nil {
	// 	log.Fatal(err.Error())
	// }
	//
	// differentFile, diffErr := dfs.Open("new3.txt")
	//
	// if diffErr != nil {
	// 	log.Fatal(diffErr.Error())
	// }
	//
	// fileRead := bufio.NewReader(differentFile)
	// line, _ := fileRead.ReadString('\n')
	// fmt.Println(line)
	//
	// line, _ = fileRead.ReadString('\n')
	// fmt.Println(line)
}
