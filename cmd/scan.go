package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gosuri/uiprogress"
	"github.com/wux1an/port-scanner"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	portOpened = 1
	portClosed = -1
	portNotest = 0
)

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

type handler struct {
	rawArgs     ConfigArgs
	scanConfig  *scanner.Config
	err         error
	output      bool
	outputFile  *os.File
	appendError error
	results     map[string]map[int]int
}

func newHandler(configArgs ConfigArgs) *handler {
	return &handler{rawArgs: configArgs}
}

func (ss *handler) handle() {
	if ss.err != nil {
		return
	}

	// create output file
	if ss.output = ss.rawArgs.OutputArg != ""; ss.output {
		ss.outputFile, ss.err = os.OpenFile(configArgs.OutputArg, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if ss.err != nil {
			color.Red("failed to create output file '%s', %v", configArgs.OutputArg, ss.err)
			return
		}
	}

	// parse args
	ss.scanConfig, ss.err = ParseConfigArgs(ss.rawArgs)
	if ss.err != nil {
		color.Red("failed to parse args, %v\n", ss.err)
		return
	}

	// init results map
	ss.results = make(map[string]map[int]int, len(ss.scanConfig.Hosts)) // ip : port : open (1 open, -1 close, 0 no test)
	for _, host := range ss.scanConfig.Hosts {
		ss.results[host.String()] = make(map[int]int, len(ss.scanConfig.Ports))
	}

	var s = scanner.NewScanner(ss.scanConfig)
	var p = s.Progress()

	var wg sync.WaitGroup
	wg.Add(2)

	// cli output
	go func() {
		defer wg.Done()

		if multi := len(ss.scanConfig.Hosts) <= 10; multi {
			ss.multiProgressBar(p)
		} else {
			ss.singleProgressBar(p)
		}

		// output to cli
		colorfulResult := ss.buildResult(ss.results)
		color.New().Println(colorfulResult)

		// output all result to file
		if configArgs.OutputArg != "" {
			color.NoColor = true
			result := ss.buildResult(ss.results)
			color.NoColor = false

			ss.appendOutputFile("\n\n" + fmt.Sprintf(result))
		}
	}()

	// scan
	go func() {
		defer wg.Done()

		s.Scan()
	}()

	wg.Wait()
}

func (ss *handler) singleProgressBar(p <-chan *scanner.ScanItem) {
	if ss.err != nil {
		return
	}

	uiprogress.Empty = ' '
	bar := uiprogress.AddBar(len(ss.scanConfig.Ports) * len(ss.scanConfig.Hosts)).
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
			ss.results[item.Host.String()][item.Port] = portOpened
			if ss.output {
				ss.appendOutputFile(fmt.Sprintf("%s:%d", item.Host, item.Port))
			}
		} else {
			ss.results[item.Host.String()][item.Port] = portClosed
		}
	}
	uiprogress.Stop()
}

func (ss *handler) multiProgressBar(p <-chan *scanner.ScanItem) {
	// initialize progress bars
	var hostBarMap = make(map[string]*uiprogress.Bar, len(ss.scanConfig.Hosts))
	for _, host := range ss.scanConfig.Hosts {
		hostBarMap[host.String()] = uiprogress.AddBar(len(ss.scanConfig.Ports)).
			AppendFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("%s (%d/%d)", b.CompletedPercentString(), b.Current(), b.Total)
			}).
			PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("%-15s  %s", host.String(), b.TimeElapsedString())
			})
	}

	uiprogress.Empty = ' '
	uiprogress.Start()

	for item := range p {
		hostBarMap[item.Host.String()].Incr()

		if item == nil {
			continue
		}

		if opened := item.Err == nil; opened {
			ss.results[item.Host.String()][item.Port] = portOpened
			if ss.output {
				ss.appendOutputFile(fmt.Sprintf("%s:%d", item.Host, item.Port))
			}
		} else {
			ss.results[item.Host.String()][item.Port] = portClosed
		}
	}
	uiprogress.Stop()
}

func (ss *handler) appendOutputFile(line string) {
	if ss.appendError != nil || ss.err != nil {
		return
	}

	_, ss.appendError = ss.outputFile.WriteString("\n" + line)
	if ss.appendError != nil {
		color.Red("failed to write result, %v", ss.err)
	} else {
		ss.outputFile.Sync()
	}
}

func (ss *handler) buildResult(results map[string]map[int]int) string {
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

	return builder.String()
}
