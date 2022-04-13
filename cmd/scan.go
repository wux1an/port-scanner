package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gosuri/uiprogress"
	"github.com/wux1an/port-scanner"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func scan() {
	config, err := ParseConfigArgs(configArgs)

	if err != nil {
		fmt.Printf("failed to parse args, %v\n", err)
		return
	}

	var s = scanner.NewScanner(config)
	var p = s.Progress()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		if multi := len(config.Hosts) <= 10; multi {
			multiProgressBar(*config, p)
		} else {
			singleProgressBar(*config, p)
		}
	}()

	go func() {
		defer wg.Done()

		s.Scan()
	}()

	wg.Wait()
}

const (
	portOpened = 1
	portClosed = -1
	portNotest = 0
)

func singleProgressBar(config scanner.Config, p <-chan *scanner.ScanItem) {
	// initialize results
	var results = make(map[string]map[int]int, len(config.Hosts)) // ip : port : open (1 open, -1 close, 0 no test)
	for _, host := range config.Hosts {
		results[host.String()] = make(map[int]int, len(config.Ports))
	}

	uiprogress.Empty = ' '
	bar := uiprogress.AddBar(len(config.Ports) * len(config.Hosts)).
		AppendFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("%s (%d/%d)", b.CompletedPercentString(), b.Current(), b.Total)
		}).
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("running  %s", b.TimeElapsedString())
		})
	uiprogress.Start()

	for item := range p {
		bar.Incr()

		if item == nil {
			continue
		}

		if opened := item.Err == nil; opened {
			results[item.Host.String()][item.Port] = portOpened
		} else {
			results[item.Host.String()][item.Port] = portClosed
		}
	}
	uiprogress.Stop()

	printResult(results)
}

func multiProgressBar(config scanner.Config, p <-chan *scanner.ScanItem) {
	// initialize progress bars
	var hostBarMap = make(map[string]*uiprogress.Bar, len(config.Hosts))
	for _, host := range config.Hosts {
		hostBarMap[host.String()] = uiprogress.AddBar(len(config.Ports)).
			AppendFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("%s (%d/%d)", b.CompletedPercentString(), b.Current(), b.Total)
			}).
			PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("%-15s  %s", host.String(), b.TimeElapsedString())
			})
	}

	uiprogress.Empty = ' '
	uiprogress.Start()

	// initialize results
	var results = make(map[string]map[int]int, len(config.Hosts)) // ip : port : open (1 open, -1 close, 0 no test)
	for _, host := range config.Hosts {
		results[host.String()] = make(map[int]int, len(config.Ports))
	}

	for item := range p {
		hostBarMap[item.Host.String()].Incr()

		if item == nil {
			continue
		}

		if opened := item.Err == nil; opened {
			results[item.Host.String()][item.Port] = portOpened
		} else {
			results[item.Host.String()][item.Port] = portClosed
		}
	}
	uiprogress.Stop()

	printResult(results)
}

var (
	cs = color.CyanString
	cb = color.New(color.FgCyan, color.Bold).Sprintf
	gs = color.GreenString
	gb = color.New(color.FgGreen, color.Bold).Sprintf
	rs = color.RedString
	rb = color.New(color.FgRed, color.Bold).Sprintf
	bs = color.BlueString
	bb = color.New(color.FgBlue, color.Bold).Sprintf
)

func printResult(results map[string]map[int]int) {
	var builder = strings.Builder{}
	for host, ports := range results {
		var closed = 0

		// get all opened ports
		var openedPorts []int
		for port, status := range ports {
			if status == portOpened {
				openedPorts = append(openedPorts, port)
			} else if status == portClosed {
				closed++
			}
		}

		if len(openedPorts) == 0 {
			continue
		}

		sort.Ints(openedPorts)

		builder.WriteString(cs("%-15s", host) + "  " +
			gb(strconv.Itoa(len(openedPorts))) + gs(" opened") + ", " +
			rb(strconv.Itoa(closed)) + rs(" closed") + ", " +
			bb(strconv.Itoa(len(ports))) + bs(" total") + cb(" => "))
		for i, p := range openedPorts {
			if sep := i != 0; sep {
				builder.WriteString(", ")
			}
			builder.WriteString(gb(strconv.Itoa(p)))
		}

		builder.WriteString("\n")
	}

	fmt.Println(builder.String())
}
