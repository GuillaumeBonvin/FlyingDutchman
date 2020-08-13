package internal

import (
	"crypto/sha1"
	"hash"
)

func CreateHash(byteStr []byte) []byte {
	var hashVal hash.Hash
	hashVal = sha1.New()
	hashVal.Write(byteStr)

	var bytes []byte

	bytes = hashVal.Sum(nil)
	return bytes
}
