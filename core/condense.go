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
	salt, tnonce, tableID := header.Salt, header.Nonce, header.TableID

	// Restore keys
	fkey, tkey := ExpandPassphrase([]byte(passphrase), salt)

	// Fetch table from IPFS and decrypt its data
	tbytes, err := ipfsContext.GetBlock(tableID)
	if err != nil {
		return "", err
	}
	tcipher, err := NewSmogCipher(tkey, tnonce)
	if err != nil {
		return "", err
	}
	var table pb.BlockTable
	if err := proto.Unmarshal(tcipher.Decrypt(tbytes), &table); err != nil {
		return "", err
	}
	fnonce, blockIDs := table.Nonce, table.Blocks

	// Fetch blocks one by one from IPFS, encrypting each one
	// and getting next block ID from it
	blocks := make([][]byte, 0)
	for _, blockID := range blockIDs {
		fmt.Printf("Retrieving %s\n", base58.Encode(blockID))
		block, err := ipfsContext.GetBlock(blockID)
		if err != nil {
			fstr := "[condense]: failed to retrieve a block: %s"
			return "", fmt.Errorf(fstr, err)
		}
		blocks = append(blocks, block)
	}

	// Concatenate blocks and return decrypted result
	fcipher, err := NewSmogCipher(fkey, fnonce)
	if err != nil {
		return "", err
	}
	return string(fcipher.Decrypt(bytes.Join(blocks, []byte{}))), nil
}
