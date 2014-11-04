package examples

import (
	"bytes"

	"github.com/willemvds/kraken"
)

type ShellCommander struct {
	cmdQueue chan []byte
	buf      bytes.Buffer
}

func (sc *ShellCommander) NextCommand() ([]byte, error) {
	cmd, ok := <-sc.cmdQueue
	if ok {
		return cmd, nil
	}
	return []byte{}, kraken.ErrNoMoreCommands
}

func (sc *ShellCommander) Write(bs []byte) (int, error) {
	return sc.buf.Write(bs)
}

func (sc *ShellCommander) AddCommand(cmd []byte) {
	sc.cmdQueue <- cmd
}

func NewShellCommander() *ShellCommander {
	sc := ShellCommander{}
	sc.cmdQueue = make(chan []byte, 20)
	return &sc
}
