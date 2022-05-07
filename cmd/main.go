package main

import (
	"os"

	"github.com/serj1c/blockchainio/blockchain"
	cli2 "github.com/serj1c/blockchainio/cli"
)

func main() {
	defer os.Exit(0)
	chain := blockchain.InitBlockChain()
	defer chain.Database.Close()

	cli := cli2.CommandLine{Blockchain: chain}
	cli.Run()
}
