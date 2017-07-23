package core

import (
	"fmt"
	"io/ioutil"

	base58 "github.com/jbenet/go-base58"
)

var (
	BlockSize = 32
	IdSize    = 34
)

// Read file, break it into blocks and publish them into IPFS network
// gradually encrypting each one with viewKey
func Vaporize(passphrase, filename string) (string, error) {
	fmtstr := "Vaporizing file '%s' with passphrase '%s'\n"
	fmt.Printf(fmtstr, filename, passphrase)

	ipfsContext, err := NewIpfsContext()
	if err != nil {
		return "", err
	}

	fileBlocks, err := getRawFileBlocks(filename)
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

func getRawFileBlocks(filename string) ([][]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var res [][]byte
	for i := 0; i < len(data); i += BlockSize {
		res = append(res, data[i:i+BlockSize])
	}
	return res, nil
}
