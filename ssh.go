package kraken

import (
	"io"

	"code.google.com/p/go.crypto/ssh"
)

const (
	_SIGINT = "\x03"
)

type sshClient struct {
	address string
	config  *ssh.ClientConfig
	ptyType string

	client  *ssh.Client
	session *ssh.Session
	Stdout  io.Reader
	Stderr  io.Reader
	Stdin   io.WriteCloser
}

func (sshclient *sshClient) Write(bs []byte) (int, error) {
	return sshclient.Stdin.Write(bs)
}

func (sshclient *sshClient) Read(buf []byte) (int, error) {
	return sshclient.Stdout.Read(buf)
}

func (sshclient *sshClient) ReadErr(buf []byte) (int, error) {
	return sshclient.Stderr.Read(buf)
}

func (sshclient *sshClient) Signal(signal ssh.Signal) error {
	return sshclient.session.Signal(signal)
}

func (sshclient *sshClient) SIGINT() {
	sshclient.Stdin.Write([]byte(_SIGINT))
}

func (sshclient *sshClient) Connect() error {
	var err error

	if sshclient.client, err = ssh.Dial("tcp", sshclient.address, sshclient.config); err != nil {
		return err
	}

	if sshclient.session, err = sshclient.client.NewSession(); err != nil {
		return err
	}

	if sshclient.Stdout, err = sshclient.session.StdoutPipe(); err != nil {
		return err
	}

	if sshclient.Stderr, err = sshclient.session.StderrPipe(); err != nil {
		return err
	}

	if sshclient.Stdin, err = sshclient.session.StdinPipe(); err != nil {
		return err
	}

	if err = sshclient.session.RequestPty(sshclient.ptyType, 80, 40, nil); err != nil {
		return err
	}

	if err = sshclient.session.Shell(); err != nil {
		return err
	}

	return nil
}

func (sshclient *sshClient) Close() error {
	return sshclient.client.Close()
}
