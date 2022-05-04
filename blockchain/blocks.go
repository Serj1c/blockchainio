package blockchain

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

type BlockChain struct {
	Blocks []*Block
}

func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{
		Data:     []byte(data),
		PrevHash: prevHash,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (ch *BlockChain) AddBlock(data string) {
	prevBlock := ch.Blocks[len(ch.Blocks)-1]
	newBlock := CreateBlock(data, prevBlock.Hash)
	ch.Blocks = append(ch.Blocks, newBlock)
}

func FirstBlock() *Block {
	return CreateBlock("FirstBlock", []byte{})
}

func InitBlockChain() *BlockChain {
	return &BlockChain{
		Blocks: []*Block{FirstBlock()},
	}
}
