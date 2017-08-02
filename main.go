package main

import (
	"fmt"
	"os"

	"github.com/creeptic/smog/core"
	"github.com/urfave/cli"
)

var (
	missingFileArgError = cli.NewExitError("FILE argument required", 1)
	missingHashArgError = cli.NewExitError("HASH argument required", 2)
	internalError       = cli.NewExitError("Internal error", 3)
)

func main() {
	app := cli.NewApp()
	vaporizeCommand := cli.Command{
		Name:    "vaporize",
		Aliases: []string{"v"},
		Usage:   "Move data to the IPFS network",
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return missingFileArgError
			}
			file := c.Args().Get(0)
			id, err := core.Vaporize(file, c.String("p"), c.Int("b"))
			if err != nil {
				return internalError
			}
			fmt.Printf("ID: %s\n", id)
			return nil
		},
	}
	condenseCommand := cli.Command{
		Name:    "condense",
		Aliases: []string{"c"},
		Usage:   "Fetch data from the IPFS network",
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return missingHashArgError
			}
			file := c.Args().Get(0)
			cs, err := core.Condense(c.String("p"), file)
			if err != nil {
				return internalError
			}
			fmt.Printf("Contents: %s\n", cs)
			return nil
		},
	}
	passphraseFlag := cli.StringFlag{
		Name:  "passphrase, p",
		Usage: "Use `PASS` to encrypt/decrypt data",
	}
	blocksizeFlag := cli.IntFlag{
		Name:  "blocksize, b",
		Usage: "Break data into blocks of `SIZE` bytes",
		Value: core.BLOCK,
	}
	app.Commands = []cli.Command{vaporizeCommand, condenseCommand}
	app.Flags = []cli.Flag{passphraseFlag, blocksizeFlag}
	app.Run(os.Args)
}
