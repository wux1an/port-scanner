package scanner

import (
	"github.com/pkg/errors"
	"net"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	Hosts           []net.IP
	Ports           []int
	Mix             bool // mix mode, disable: scan queue like 'host1:port1', 'host1:port2', enable: 'host1:port1', 'host2:port1'
	Thread          int
	TimeoutInSecond int
}

type ScanItem struct {
	Host net.IP
	Port int
	Err  error
}

type Scanner struct {
	config   *Config
	progress chan *ScanItem
}

func NewScanner(config *Config) (*Scanner, error) {

	if len(config.Hosts) == 0 {
		return nil, errors.New("no Hosts specified")
	}
	if len(config.Ports) == 0 {
		return nil, errors.New("no Ports specified")
	}
	if config.Thread < 1 {
		return nil, errors.New("thread number is less than 1")
	}
	if config.TimeoutInSecond <= 0 {
		return nil, errors.New("timeout is less than 0")
	}

	return &Scanner{
		config:   config,
		progress: make(chan *ScanItem),
	}, nil
}

func (s *Scanner) Progress() <-chan *ScanItem {
	return s.progress
}

func (s Scanner) Scan() {
	var wg sync.WaitGroup
	var ch = make(chan interface{}, s.config.Thread)

	if s.config.Mix {
		for _, port := range s.config.Ports {
			for _, host := range s.config.Hosts {
				s.scan0(&ScanItem{
					Host: host,
					Port: port,
				}, &wg, ch)
			}
		}
	} else {
		for _, host := range s.config.Hosts {
			for _, port := range s.config.Ports {
				s.scan0(&ScanItem{
					Host: host,
					Port: port,
				}, &wg, ch)
			}
		}
	}

	wg.Wait()

	close(s.progress)
}

func (s Scanner) scan0(item *ScanItem, wg *sync.WaitGroup, ch chan interface{}) {
	ch <- nil
	wg.Add(1)
	go func() {
		defer func() {
			s.progress <- item
			wg.Done()
			<-ch
		}()

		_, item.Err = net.DialTimeout("tcp",
			net.JoinHostPort(item.Host.String(), strconv.Itoa(item.Port)),
			time.Duration(s.config.TimeoutInSecond)*time.Second,
		)
	}()
}
