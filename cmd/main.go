package main

import (
	"fmt"
	"github.com/serj1c/blockchainio/blockchain"
	"strconv"
)

func main() {
	chain := blockchain.InitBlockChain()

	chain.AddBlock("second block")
	chain.AddBlock("third block")

	for _, block := range chain.Blocks {
		fmt.Printf("Prev hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println() // for spacing
	}
}
