package scanner

import (
	"github.com/pkg/errors"
	"net"
)

type Config struct {
	hosts           []net.IP
	ports           []int
	thread          int
	timeoutInSecond int
}

// NewConfig timeoutInSecond is the timeout of tcp connect
func NewConfig(hosts []net.IP, ports []int, thread int, timeoutInSecond int) (*Config, error) {
	if len(hosts) == 0 {
		return nil, errors.New("no hosts specified")
	}
	if len(ports) == 0 {
		return nil, errors.New("no ports specified")
	}
	if thread < 1 {
		return nil, errors.New("thread number is less than 1")
	}
	if timeoutInSecond <= 0 {
		return nil, errors.New("timeout is less than 0")
	}

	return &Config{hosts: hosts, ports: ports, thread: thread, timeoutInSecond: timeoutInSecond}, nil
}

type Scanner struct {
	config *Config
}

func NewScanner(config *Config) (*Scanner, error) {
	// todo implement

	return nil, nil
}
