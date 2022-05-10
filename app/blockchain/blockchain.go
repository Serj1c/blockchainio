package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type Iterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DbExists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		firstBlock := FirstBlock(cbtx)
		fmt.Println("First block created")
		err = txn.Set(firstBlock.Hash, firstBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = txn.Set([]byte("lh"), firstBlock.Hash)
		lastHash = firstBlock.Hash
		return err

	})
	if err != nil {
		log.Panic(err)
	}

	return &BlockChain{
		LastHash: lastHash,
		Database: db,
	}
}

func ContinueBlockChain(address string) *BlockChain {
	if DbExists() == false {
		fmt.Println("No existing blockchain found")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		lastHash, err = item.ValueCopy(lastHash)

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	chain := BlockChain{
		LastHash: lastHash,
		Database: db,
	}
	return &chain
}

func (ch *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction

	spentTxOs := make(map[string][]int)

	iter := ch.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txId := hex.EncodeToString(tx.Id)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTxOs[txId] != nil {
					for _, spentOut := range spentTxOs[txId] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					if in.UsesKey(pubKeyHash) {
						inTxId := hex.EncodeToString(in.Id)
						spentTxOs[inTxId] = append(spentTxOs[inTxId], in.Out)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTxs
}

func (ch *BlockChain) FindUTxO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := ch.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (ch *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := ch.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txId := hex.EncodeToString(tx.Id)

		for outIdx, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txId] = append(unspentOuts[txId], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}

func (ch *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	err := ch.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		lastHash, err = item.ValueCopy(lastHash)
		return err
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := CreateBlock(transactions, lastHash)

	err = ch.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("lh"), newBlock.Hash)

		ch.LastHash = newBlock.Hash

		return err
	})

	if err != nil {
		log.Panic(err)
	}
}

func DbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func (ch *BlockChain) Iterator() *Iterator {
	return &Iterator{
		CurrentHash: ch.LastHash,
		Database:    ch.Database,
	}
}

func (it *Iterator) Next() *Block {
	var block *Block

	err := it.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(it.CurrentHash)
		if err != nil {
			log.Panic(err)
		}
		encodedBlock, err := item.ValueCopy(nil)
		block = Deserialize(encodedBlock)
		return err
	})
	if err != nil {
		log.Panic(err)
	}

	it.CurrentHash = block.PrevHash

	return block
}

func (ch *BlockChain) FindTransaction(Id []byte) (Transaction, error) {
	iterator := ch.Iterator()

	for {
		block := iterator.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.Id, Id) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("transaction does not exist")
}

func (ch *BlockChain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	prevTxs := make(map[string]Transaction)

	for _, input := range tx.Inputs {
		prevTx, err := ch.FindTransaction(input.Id)
		if err != nil {
			log.Panic(err)
		}
		prevTxs[hex.EncodeToString(prevTx.Id)] = prevTx
	}

	tx.Sign(privateKey, prevTxs)
}

func (ch *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTxs := make(map[string]Transaction)

	for _, input := range tx.Inputs {
		prevTx, err := ch.FindTransaction(input.Id)
		if err != nil {
			log.Panic(err)
		}
		prevTxs[hex.EncodeToString(prevTx.Id)] = prevTx
	}

	return tx.Verify(prevTxs)
}
