package core

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/creeptic/smog/pb"
	"github.com/gogo/protobuf/proto"
	base58 "github.com/jbenet/go-base58"
)

const (
	BLOCK = 1024
	ID    = 34
	KEY   = 32
	NONCE = 16
	SALT  = 32
)

// Read file, break it into blocks and publish them into IPFS network
func Vaporize(filename, passphrase string, blocksize int) (string, error) {
	// Initialize IPFS connection
	ipfsContext, err := NewIpfsContext()
	if err != nil {
		return "", err
	}

	// Generate required cryptographic values
	buf := make([]byte, SALT+2*NONCE)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	salt, fnonce, tnonce := buf[:SALT], buf[SALT:SALT+NONCE], buf[SALT+NONCE:]
	fkey, tkey := ExpandPassphrase([]byte(passphrase), salt)

	// Read raw data
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// Encrypt and store raw data
	ftable, err := vaporizeData(ipfsContext, fkey, fnonce, data, blocksize)
	if err != nil {
		return "", err
	}

	// Encrypt and store table in IPFS
	table := &pb.BlockTable{Nonce: fnonce, Blocks: ftable}
	tableID, err := vaporizeTable(ipfsContext, tkey, tnonce, table)
	if err != nil {
		return "", err
	}

	// Store header in the IPFS, return its id as a link to data
	header := &pb.Header{Salt: salt, Nonce: tnonce, TableID: tableID}
	headerID, err := vaporizeHeader(ipfsContext, header)
	if err != nil {
		return "", err
	}
	return base58.Encode(headerID), nil
}

// Store raw file data in IPFS, return block id index
func vaporizeData(ipfs IpfsContext, fkey, fnonce, data []byte, bs int) ([][]byte, error) {
	fcipher, err := NewSmogCipher(fkey, fnonce)
	if err != nil {
		return nil, err
	}
	// Store blocks in IPFS
	ftable := make([][]byte, 0)
	for _, block := range chunks(fcipher.Encrypt(data), BLOCK) {
		blockID, err := ipfs.PutBlock(block)
		if err != nil {
			fstr := "[vaporize]: failed to put a new block: %s"
			return nil, errors.New(fmt.Sprintf(fstr, err))
		}
		ftable = append(ftable, blockID)
	}
	return ftable, nil
}

// Store block index in IPFS, return its ID
func vaporizeTable(ipfs IpfsContext, tkey, tnonce []byte, table *pb.BlockTable) ([]byte, error) {
	tableBytes, err := proto.Marshal(table)
	if err != nil {
		return nil, err
	}
	tcipher, err := NewSmogCipher(tkey, tnonce)
	if err != nil {
		return nil, err
	}
	tableID, err := ipfs.PutBlock(tcipher.Encrypt(tableBytes))
	if err != nil {
		return nil, err
	}
	return tableID, nil
}

// Store header in IPFS, return its ID
func vaporizeHeader(ipfs IpfsContext, header *pb.Header) ([]byte, error) {
	headerBytes, err := proto.Marshal(header)
	if err != nil {
		return nil, err
	}
	headerID, err := ipfs.PutBlock(headerBytes)
	if err != nil {
		return nil, err
	}
	return headerID, nil
}

// Break 'data' byte slice into chunks of size 'blocksize'
func chunks(data []byte, blocksize int) [][]byte {
	var res [][]byte
	length := len(data)
	for i := 0; i < length; i += blocksize {
		end := i + blocksize
		if end > length {
			end = length
		}
		res = append(res, data[i:end])
	}
	return res
}
