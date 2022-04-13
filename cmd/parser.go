package main

import (
	"fmt"
	"github.com/malfunkt/iprange"
	"github.com/pkg/errors"
	"github.com/wux1an/port-scanner"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func ParseConfigArgs(configArgs ConfigArgs) (*scanner.Config, error) {
	var config scanner.Config

	if configArgs.ThreadArg <= 0 {
		return nil, errors.New(fmt.Sprintf("invalid thread number '%d'", configArgs.ThreadArg))
	}

	if configArgs.TimeoutInSecondArg < 0 {
		return nil, errors.New(fmt.Sprintf("invalid timeout in second '%d'", configArgs.TimeoutInSecondArg))
	}

	if configArgs.HostsArg == "" {
		return nil, errors.New("no hosts specified")
	}

	if configArgs.PortsArg == "" {
		return nil, errors.New("no ports specified")
	}

	// parse hosts
	t, err := iprange.ParseList(configArgs.HostsArg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse hosts")
	}
	config.Hosts = t.Expand()

	// parse ports
	config.Ports, err = parsePortRange(configArgs.PortsArg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse ports")
	}

	// shuffle if needed
	if configArgs.Shuffle {
		rand.Shuffle(len(config.Hosts), func(i, j int) {
			config.Hosts[i], config.Hosts[j] = config.Hosts[j], config.Hosts[i]
		})
		rand.Shuffle(len(config.Ports), func(i, j int) {
			config.Ports[i], config.Ports[j] = config.Ports[j], config.Ports[i]
		})
	}

	config.Thread = configArgs.ThreadArg
	config.TimeoutInSecond = configArgs.TimeoutInSecondArg
	config.Mix = configArgs.Shuffle

	return &config, nil
}

var (
	regPortRange = regexp.MustCompile(`^(\d+(-\d+)*)(,\d+(-\d+)*)*$`)
	maxPort      = 65535
	minPort      = 0
)

// parsePortRange, example: 22 1-1024 22,234 22,234,1024-2048
func parsePortRange(in string) ([]int, error) {
	if !regPortRange.MatchString(in) {
		return nil, errors.New("syntax error")
	}

	var portMap = make(map[int]bool, maxPort-minPort+1)
	var ps = strings.Split(in, ",")
	for _, p := range ps {
		if strings.ContainsRune(p, '-') { // range
			var pp = strings.Split(p, "-")
			var min, _ = strconv.Atoi(pp[0])
			var max, _ = strconv.Atoi(pp[1])
			if min > max {
				return nil, errors.New(fmt.Sprintf("failed to parse '%s', the minimum port is greater than the maximum", p))
			}

			for i := min; i <= max; i++ {
				if i > maxPort || i < minPort {
					return nil, errors.New(fmt.Sprintf("the port number '%d' is invalid exists in '%s'", i, p))
				}
				portMap[i] = true
			}
		} else {
			var i, _ = strconv.Atoi(p)
			if i > maxPort || i < minPort {
				return nil, errors.New(fmt.Sprintf("the port number '%d' is invalid exists in '%s'", i, in))
			}
			portMap[i] = true
		}
	}

	var result = make([]int, 0, maxPort-minPort+1)
	for port, scan := range portMap {
		if scan {
			result = append(result, port)
		}
	}
	sort.Ints(result)
	return result, nil
}
