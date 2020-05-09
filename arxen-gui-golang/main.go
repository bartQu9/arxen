package main

import (
	log "github.com/sirupsen/logrus"
	"main/client"
	"main/serverhandler"
	"os"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}


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
