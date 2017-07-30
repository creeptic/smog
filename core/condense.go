package core

import (
	"bytes"
	"fmt"

	"github.com/creeptic/smog/pb"
	"github.com/gogo/protobuf/proto"
	base58 "github.com/jbenet/go-base58"
)

// Gradually fetch raw blocks from IPFS network and
// decrypt them into a file
func Condense(passphrase, headId string) (string, error) {
	// Initialize IPFS connection
	ipfsContext, err := NewIpfsContext()
	if err != nil {
		return "", err
	}

	// Get header from IPFS
	headerBytes, err := ipfsContext.GetBlock(base58.Decode(headId))
	if err != nil {
		return "", err
	}
	var header pb.Header
	if err := proto.Unmarshal(headerBytes, &header); err != nil {
		return "", err
	}

	// Expand passphrase
	cipher, err := NewSmogCipher(base58.Decode(passphrase), header.Salt, header.Nonce)
	if err != nil {
		return "", err
	}

	//
	var blocks [][]byte
	nullBlockId, nextBlockId := make([]byte, IdSize), header.Head
	for !bytes.Equal(nextBlockId, nullBlockId) {
		fmt.Printf("Retrieving %s\n", base58.Encode(nextBlockId))
		securedblock, err := ipfsContext.GetBlock(nextBlockId)
		if err != nil {
			fstr := "[condense]: failed to retrieve a block: %s"
			return "", fmt.Errorf(fstr, err)
		}
		markedBlock := cipher.Decrypt(securedblock)
		var rawBlock []byte
		nextBlockId, rawBlock = markedBlock[:IdSize], markedBlock[IdSize:]
		blocks = append(blocks, rawBlock)
	}
	// Wait what? Really? It's 2017!
	for i, j := 0, len(blocks)-1; i < j; i, j = i+1, j-1 {
		blocks[i], blocks[j] = blocks[j], blocks[i]
	}
	return string(bytes.Join(blocks, []byte{})), nil
}
