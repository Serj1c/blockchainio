package blockchain

import (
	"bytes"
	"encoding/binary"
	"log"
)

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, num); err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
