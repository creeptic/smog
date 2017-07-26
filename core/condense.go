package core

import (
	"bytes"
	"fmt"

	base58 "github.com/jbenet/go-base58"
)

// Gradually fetch raw blocks from IPFS network and
// decrypt them into a file
func Condense(passphrase, headId string) (string, error) {
	ipfsContext, err := NewIpfsContext()
	if err != nil {
		return "", err
	}

	var blocks [][]byte
	nullBlockId, nextBlockId := make([]byte, IdSize), base58.Decode(headId)
	for !bytes.Equal(nextBlockId, nullBlockId) {
		fmt.Printf("Retrieving %s\n", base58.Encode(nextBlockId))
		block, err := ipfsContext.GetBlock(nextBlockId)
		if err != nil {
			fstr := "[condense]: failed to retrieve a block: %s"
			return "", fmt.Errorf(fstr, err)
		}
		var rawBlock []byte
		nextBlockId, rawBlock = block[:IdSize], block[IdSize:]
		blocks = append(blocks, rawBlock)
	}
	// Wait what? Really? It's 2017!
	for i, j := 0, len(blocks)-1; i < j; i, j = i+1, j-1 {
		blocks[i], blocks[j] = blocks[j], blocks[i]
	}
	return string(bytes.Join(blocks, []byte{})), nil
}
