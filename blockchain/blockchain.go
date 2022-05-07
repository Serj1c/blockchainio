package blockchain

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
)

const dbPath = "./tmp/blocks"

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type Iterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No blockchain found")
			firstBlock := FirstBlock()
			fmt.Println("First block proved")
			err = txn.Set(firstBlock.Hash, firstBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = txn.Set([]byte("lh"), firstBlock.Hash)
			if err != nil {
				return err
			}
			lastHash = firstBlock.Hash
		} else {
			item, err := txn.Get([]byte("lh"))
			if err != nil {
				log.Panic(err)
			}
			lastHash, err = item.ValueCopy(lastHash)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return &BlockChain{
		LastHash: lastHash,
		Database: db,
	}
}

func (ch *BlockChain) AddBlock(data string) {
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

	newBlock := CreateBlock(data, lastHash)

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
