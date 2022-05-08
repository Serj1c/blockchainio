package main

import (
	"os"

	"github.com/serj1c/blockchainio/app/cli"
)

func main() {
	defer os.Exit(0)

	commandLine := cli.CommandLine{}
	commandLine.Run()
}
