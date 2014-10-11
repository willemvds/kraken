package kraken

import (
	"errors"
	"io"

	"code.google.com/p/go.crypto/ssh"
)

var ErrNoMoreCommands = errors.New("No More Commands")

type jobStatus int

const (
	JOB_NO_MORE_COMMANDS jobStatus = iota
	JOB_REMOTE_CONNECTION_CLOSED
)

// Commander will receive terminal output and be asked what to do next
// by calling NextCommand until ErrNoMoreCommands is returned.
type Commander interface {
	NextCommand() ([]byte, error)
	io.Writer
}

type Job struct {
	sshclient  sshClient
	commander  Commander
	statusChan chan jobStatus
}

func NewJob(addr string, conf *ssh.ClientConfig, c Commander) *Job {
	job := Job{}
	job.sshclient = sshClient{address: addr, config: conf}
	job.commander = c
	return &job
}

// Start connects using the provided SSH details and reads/writes
// data over the connection. It also returns a channel to read
// job status messages from.
func (job *Job) Start() (<-chan jobStatus, error) {
	if err := job.sshclient.Connect(); err != nil {
		return nil, err
	}
	if err := job.startCommandLoop(); err != nil {
		return nil, err
	}
	job.statusChan = make(chan jobStatus)
	return job.statusChan, nil
}

func (job *Job) startCommandLoop() error {
	go func() {
		buffer := make([]byte, 100)
		for n, err := job.sshclient.Read(buffer); err == nil || n > 0; n, err = job.sshclient.Read(buffer) {
			_, cerr := job.commander.Write(buffer[:n])
			if cerr != nil {
				print(cerr)
			}
		}
		job.statusChan <- JOB_REMOTE_CONNECTION_CLOSED
	}()
	go func() {
		for command, err := job.commander.NextCommand(); err != ErrNoMoreCommands; command, err = job.commander.NextCommand() {
			job.sshclient.Write(command)
		}
		job.statusChan <- JOB_NO_MORE_COMMANDS
	}()
	return nil
}

// Complete is used to tell the job that it can perform cleanup
// like closing the shh connection.
func (job *Job) Complete() error {
	return job.sshclient.Close()
}
