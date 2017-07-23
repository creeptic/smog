package core

import (
	"github.com/ipfs/go-ipfs-api"
	base58 "github.com/jbenet/go-base58"
)

var (
	defaultRepoPath = "~/.ipfs"
	defaultAPIAddr  = "/ip4/127.0.0.1/tcp/5001"
)

type IpfsContext interface {
	GetBlock(id []byte) ([]byte, error)
	PutBlock(block []byte) ([]byte, error)
}

func NewIpfsContext() (IpfsContext, error) {
	return &ipfsContext{shell: shell.NewShell(defaultAPIAddr)}, nil
}

type ipfsContext struct {
	shell *shell.Shell
}

func (ctx ipfsContext) GetBlock(id []byte) ([]byte, error) {
	return ctx.shell.BlockGet(base58.Encode(id))
}

func (ctx ipfsContext) PutBlock(block []byte) ([]byte, error) {
	id, err := ctx.shell.BlockPut(block)
	return base58.Decode(id), err
}
