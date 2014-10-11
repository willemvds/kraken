package kraken

import (
	"errors"
	"io"

	"code.google.com/p/go.crypto/ssh"
)

var ErrStop = errors.New("Stop Executing")

type jobStatus int

const (
	JOB_COMPLETED jobStatus = iota
	JOB_REMOTE_CONNECTION_CLOSED
)

type commander interface {
	Command() ([]byte, error)
	io.Writer
}

type Job struct {
	sshclient  sshClient
	commander  commander
	statusChan chan jobStatus
}

func NewJob(addr string, conf *ssh.ClientConfig, c commander) *Job {
	job := Job{}
	job.sshclient = sshClient{address: addr, config: conf}
	job.commander = c
	return &job
}

func (job *Job) Start() (<-chan jobStatus, error) {
	if err := job.sshclient.Connect(); err != nil {
		return nil, err
	}
	if err := job.StartCommandLoop(); err != nil {
		return nil, err
	}
	job.statusChan = make(chan jobStatus)
	return job.statusChan, nil
}

func (job *Job) StartCommandLoop() error {
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
		for command, err := job.commander.Command(); err != ErrStop; command, err = job.commander.Command() {
			job.sshclient.Write(command)
		}
		job.statusChan <- JOB_COMPLETED
	}()
	return nil
}
