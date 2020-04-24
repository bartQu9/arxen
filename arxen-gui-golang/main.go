package main

import (
	"main/client"
)

func main() {
	cli := client.NewClient()

	go cli.HttpServer()

	cli.TestSetup()

}
