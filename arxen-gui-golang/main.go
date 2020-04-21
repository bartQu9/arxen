package main

import "main/client"

func main() {
	cli := client.NewClient()

	cli.HttpServer()
}
