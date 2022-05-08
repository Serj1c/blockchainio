package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	Id      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxInput struct {
	Id  []byte
	Out int
	Sig string
}

type TxOutput struct {
	Value  int
	PubKey string
}

func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for id, outs := range validOutputs {
		txID, err := hex.DecodeString(id)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TxInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TxOutput{amount, to})

	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetId()

	return &tx
}

func (tx *Transaction) SetId() {
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.Id = hash[:]
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txin := TxInput{[]byte{}, -1, data}
	txout := TxOutput{100, to}

	tx := Transaction{
		Id:      nil,
		Inputs:  []TxInput{txin},
		Outputs: []TxOutput{txout},
	}

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].Id) == 0 && tx.Inputs[0].Out == -1
}

func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}