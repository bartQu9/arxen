package main

import (
	"log"
	"main/client"
	"main/serverhandler"
)

func main() {
	cli := client.NewClient()

	s, err := serverhandler.NewClientServer(cli)
	if err != nil {
		log.Fatal(err)
	}

	cli.TestSetup()

	err = s.Serve(8085)
	if err != nil {
		log.Fatal(err)
	}
}
