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
	var hosts = t.Expand()

	// parse ports
	ports, err := parsePortRange(configArgs.PortsArg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse ports")
	}

	// shuffle if needed
	if configArgs.Shuffle {
		rand.Shuffle(len(hosts), func(i, j int) {
			hosts[i], hosts[j] = hosts[j], hosts[i]
		})
		rand.Shuffle(len(ports), func(i, j int) {
			ports[i], ports[j] = ports[j], ports[i]
		})
	}

	config, _ := scanner.NewConfig(hosts, ports, configArgs.ThreadArg, configArgs.TimeoutInSecondArg, configArgs.Shuffle)
	return config, nil
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
