package core

import (
	"fmt"

	base58 "github.com/jbenet/go-base58"
)

var (
	BlockSize = 32
	IdSize    = 34
)

// Read file, break it into blocks and publish them into IPFS network
// gradually encrypting each one with viewKey
func Vaporize(passphrase, filename string) (string, error) {
	ipfsContext, err := NewIpfsContext()
	if err != nil {
		return "", err
	}

	fileBlocks, err := getFileBlocks(filename)
	if err != nil {
		return "", err
	}
	nextBlockId := make([]byte, IdSize)
	for _, rawBlock := range fileBlocks {
		markedBlock := append(nextBlockId, rawBlock...)
		nextBlockId, err = ipfsContext.PutBlock(markedBlock)
		fmt.Printf("Storing %s\n\n", base58.Encode(nextBlockId))
		if err != nil {
			fstr := "[vaporize]: failed to put a new block: %s"
			return "", fmt.Errorf(fstr, err)
		}
	}
	return base58.Encode(nextBlockId), nil
}
