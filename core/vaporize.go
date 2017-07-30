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
	BlockSize = 32
	IdSize    = 34
	KeySize   = 32
	NonceSize = 16
	SaltSize  = 32
)

// Read file, break it into blocks and publish them into IPFS network
// gradually encrypting each one with viewKey
func Vaporize(passphrase, filename string) (string, error) {
	// Initialize IPFS connection
	ipfsContext, err := NewIpfsContext()
	if err != nil {
		return "", err
	}
	// Read raw data
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// Break file data into blocks
	fileBlocks := getBlocks(data, BlockSize)

	// Generate required randomness
	salt, nonce, err := generateRandomness()
	if err != nil {
		return "", err
	}

	// Expand passphrase into a suitable encryption key
	cipher, err := NewSmogCipher(base58.Decode(passphrase), salt, nonce)
	if err != nil {
		return "", err
	}

	// Store blocks in IPFS chaining them by IDs and encrypting
	nextBlockId := make([]byte, IdSize)
	for _, rawBlock := range fileBlocks {
		markedBlock := append(nextBlockId, rawBlock...)
		securedBlock := cipher.Encrypt(markedBlock)
		nextBlockId, err = ipfsContext.PutBlock(securedBlock)
		if err != nil {
			fstr := "[vaporize]: failed to put a new block: %s"
			return "", fmt.Errorf(fstr, err)
		}
		fmt.Printf("Storing %s\n\n", base58.Encode(nextBlockId))
	}

	// Store header in IPFS, return its ID as a link to the data
	header := pb.Header{Salt: salt, Nonce: nonce, Head: nextBlockId}
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

func generateRandomness() ([]byte, []byte, error) {
	buf := make([]byte, SaltSize+NonceSize)
	if _, err := rand.Read(buf); err != nil {
		return nil, nil, err
	}
	return buf[:SaltSize], buf[SaltSize:], nil
}

func getBlocks(data []byte, blocksize int) [][]byte {
	var res [][]byte
	for i := 0; i < len(data); i += blocksize {
		res = append(res, data[i:i+blocksize])
	}
	return res
}
