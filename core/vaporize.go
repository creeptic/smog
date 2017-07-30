package core

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"

	"github.com/creeptic/smog/pb"
	"github.com/gogo/protobuf/proto"
	base58 "github.com/jbenet/go-base58"
)

var (
	BLOCK = 32
	ID    = 34
	KEY   = 32
	NONCE = 16
	SALT  = 32
)

// Read file, break it into blocks and publish them into IPFS network
// gradually encrypting each one with viewKey
func Vaporize(passphrase, filename string) (string, error) {
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

	// Expand passphrase into a suitable encryption key
	fcipher, err := NewSmogCipher(fkey, fnonce)
	if err != nil {
		return "", err
	}

	// Read raw data
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// Store blocks in IPFS chaining them by IDs and encrypting
	ftable := make([][]byte, 0)
	for _, block := range getBlocks(fcipher.Encrypt(data), BLOCK) {
		blockID, err := ipfsContext.PutBlock(block)
		if err != nil {
			fstr := "[vaporize]: failed to put a new block: %s"
			return "", fmt.Errorf(fstr, err)
		}
		ftable = append(ftable, blockID)
		fmt.Printf("Storing %s\n\n", base58.Encode(blockID))
	}

	// Encrypt and store table in IPFS
	table := pb.BlockTable{Nonce: fnonce, Blocks: ftable}
	tableBytes, err := proto.Marshal(&table)
	if err != nil {
		return "", err
	}
	tcipher, err := NewSmogCipher(tkey, tnonce)
	if err != nil {
		return "", err
	}
	tableID, err := ipfsContext.PutBlock(tcipher.Encrypt(tableBytes))
	if err != nil {
		return "", err
	}

	// Store header in IPFS, return its ID as a link to the data
	header := pb.Header{Salt: salt, Nonce: tnonce, TableID: tableID}
	headerBytes, err := proto.Marshal(&header)
	if err != nil {
		return "", err
	}
	headerID, err := ipfsContext.PutBlock(headerBytes)
	if err != nil {
		return "", err
	}
	return base58.Encode(headerID), nil
}

func getBlocks(data []byte, blocksize int) [][]byte {
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
