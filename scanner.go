package scanner

import (
	"github.com/pkg/errors"
	"net"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	hosts           []net.IP
	ports           []int
	mix             bool // mix mode, disable: scan queue like 'host1:port1', 'host1:port2', enable: 'host1:port1', 'host2:port1'
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

type ScanItem struct {
	host net.IP
	port int
	err  error
}

type Scanner struct {
	config   *Config
	progress chan *ScanItem
}

func NewScanner(config *Config) *Scanner {
	return &Scanner{
		config:   config,
		progress: make(chan *ScanItem),
	}
}

func (s *Scanner) Progress() chan<- *ScanItem {
	return s.progress
}

func (s Scanner) Scan() {
	if s.config.mix {
		for _, port := range s.config.ports {
			for _, host := range s.config.hosts {
				s.Scan0(&ScanItem{
					host: host,
					port: port,
				})
			}
		}
	} else {
		for _, host := range s.config.hosts {
			for _, port := range s.config.ports {
				s.Scan0(&ScanItem{
					host: host,
					port: port,
				})
			}
		}
	}
}

func (s Scanner) Scan0(item *ScanItem) {
	var wg sync.WaitGroup
	var ch = make(chan interface{}, s.config.thread)

	ch <- nil
	wg.Add(1)
	go func() {
		defer func() {
			<-ch
			wg.Done()
		}()

		_, item.err = net.DialTimeout("tcp",
			net.JoinHostPort(item.host.String(), strconv.Itoa(item.port)),
			time.Duration(s.config.timeoutInSecond)*time.Second,
		)
	}()

	wg.Wait()
	close(s.progress)
}
