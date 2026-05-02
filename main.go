package main

import (
	"kyrux/core/bootstrap"
	"kyrux/core/cli"
	"log"
	"os"

	_ "kyrux/core/apps"
)

func main() {
	if len(os.Args) > 1 {
		cli.Run(os.Args[1:])
		return
	}

	fw, err := bootstrap.Init(".env")
	if err != nil {
		log.Fatal(err)
	}

	if err := fw.Run(); err != nil {
		log.Fatal(err)
	}
}
