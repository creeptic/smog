package core

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/creeptic/smog/pb"
	"github.com/gogo/protobuf/proto"
	base58 "github.com/jbenet/go-base58"
)

// Gradually fetch raw blocks from IPFS network and
// decrypt them into a file
func Condense(passphrase, head string) (string, error) {
	// Initialize IPFS connection
	ipfs, err := NewIpfsContext()
	if err != nil {
		return "", err
	}

	// Get header from IPFS
	header, err := condenseHeader(ipfs, base58.Decode(head))
	if err != nil {
		return "", err
	}

	// Restore keys
	fkey, tkey := ExpandPassphrase([]byte(passphrase), header.Salt)

	// Get table from IPFS
	table, err := condenseTable(ipfs, tkey, header)
	if err != nil {
		return "", err
	}

	// Fetch and decrypt blocks
	data, err := condenseData(ipfs, fkey, table)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func condenseHeader(ipfs IpfsContext, headID []byte) (*pb.Header, error) {
	headerBytes, err := ipfs.GetBlock(headID)
	if err != nil {
		return nil, err
	}
	var header pb.Header
	if err := proto.Unmarshal(headerBytes, &header); err != nil {
		return nil, err
	}
	return &header, nil
}

// Fetch table from IPFS and decrypt its data
func condenseTable(ipfs IpfsContext, tkey []byte, header *pb.Header) (*pb.BlockTable, error) {
	tbytes, err := ipfs.GetBlock(header.TableID)
	if err != nil {
		return nil, err
	}
	tcipher, err := NewSmogCipher(tkey, header.Nonce)
	if err != nil {
		return nil, err
	}
	var table pb.BlockTable
	if err := proto.Unmarshal(tcipher.Decrypt(tbytes), &table); err != nil {
		return nil, err
	}
	return &table, nil
}

// Fetch blocks from IPFS and return decryption of their concatenation
func condenseData(ipfs IpfsContext, fkey []byte, table *pb.BlockTable) ([]byte, error) {
	blocks := make([][]byte, 0)
	for _, blockID := range table.Blocks {
		block, err := ipfs.GetBlock(blockID)
		if err != nil {
			fstr := "[condense]: failed to retrieve a block: %s"
			return nil, errors.New(fmt.Sprintf(fstr, err))
		}
		blocks = append(blocks, block)
	}
	fcipher, err := NewSmogCipher(fkey, table.Nonce)
	if err != nil {
		return nil, err
	}
	return fcipher.Decrypt(bytes.Join(blocks, []byte{})), nil
}
