package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"code.google.com/p/go.crypto/ssh"
	"github.com/willemvds/kraken"
	"github.com/willemvds/kraken/examples"
)

var user string
var host string
var keypath string

func init() {
	flag.StringVar(&user, "user", "", "user on remote machine")
	flag.StringVar(&host, "host", "127.0.0.1:22", "ip:port of the remote machine")
	flag.StringVar(&keypath, "key", "", "path to ssh key to use")
}

func main() {
	flag.Parse()

	cmdr := examples.NewShellCommander()

	keybytes, err := ioutil.ReadFile(keypath)
	if err != nil {
		fmt.Println("failed to read key file:", err)
		return
	}
	signer, err := ssh.ParsePrivateKey(keybytes)
	if err != nil {
		fmt.Println("failed to parse key:", err)
		return
	}
	cfg := ssh.ClientConfig{}
	cfg.User = user
	cfg.Auth = []ssh.AuthMethod{
		ssh.PublicKeys(signer),
	}
	job := kraken.NewJob(host, &cfg, cmdr)

	statusChan, err := job.Start()
	if err != nil {
		fmt.Println("job failed to start:", err)
		return
	}

	go func() {
		for {
			line, err := cmdr.Buf.ReadString('\n')
			if err != nil && err != io.EOF {
				fmt.Println("[remote] err", err)
			}
			if len(line) > 0 {
				fmt.Printf("[remote] %s", line)
			}
		}
	}()

	go func() {
		for {
			in := bufio.NewReader(os.Stdin)
			line, err := in.ReadBytes('\n')
			if err != nil {
				continue
			}
			cmdr.AddCommand(line)
		}
	}()

	for {
		status := <-statusChan
		if status == kraken.JOB_REMOTE_CONNECTION_CLOSED {
			fmt.Println("Remote connection closed, bye.")
			job.Complete()
			break
		}
	}
}
