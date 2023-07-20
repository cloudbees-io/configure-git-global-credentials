package main

import (
	"log"

	"github.com/cloudbees-io/configure-git-global-credentials/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
