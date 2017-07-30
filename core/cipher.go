package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

// Key expansion
func ExpandPassphrase(passphraseBytes, salt []byte) ([]byte, []byte) {
	bytes := pbkdf2.Key(passphraseBytes, salt, 4096, 2*KEY, sha256.New)
	return bytes[:KEY], bytes[KEY:]
}

// Encryption interface
type SmogCipher interface {
	Encrypt(ptext []byte) []byte
	Decrypt(ctext []byte) []byte
}

type smogCipher struct {
	stream cipher.Stream
}

func NewSmogCipher(key, nonce []byte) (SmogCipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return smogCipher{cipher.NewCTR(block, nonce)}, nil
}

func (c smogCipher) Encrypt(ptext []byte) []byte {
	ctext := make([]byte, len(ptext))
	c.stream.XORKeyStream(ctext, ptext)
	return ctext
}

func (c smogCipher) Decrypt(ctext []byte) []byte {
	ptext := make([]byte, len(ctext))
	c.stream.XORKeyStream(ptext, ctext)
	return ptext
}
