package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gosuri/uiprogress"
	"github.com/wux1an/port-scanner"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
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
	start       time.Time
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

	var s, _ = scanner.NewScanner(ss.scanConfig)
	var p = s.Progress()

	var wg sync.WaitGroup
	wg.Add(2)

	// ctrl+c handler
	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		uiprogress.Stop()
		fmt.Println("canceled by ctrl+c")
		color.New().Println(ss.buildResult(ss.results))
		os.Exit(0)
	}()

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
			var t = color.NoColor
			color.NoColor = true
			result := ss.buildResult(ss.results)
			color.NoColor = t

			ss.appendOutputFile("\n\n" + fmt.Sprintf(result))
		}
	}()

	// scan
	go func() {
		defer wg.Done()

		ss.start = time.Now()
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
	/*
		=================================================

		Total:  5 hosts  x  500 ports
		Start:  2022-04-14 10:31:51  Cost: 25.2620579s
		Finish: 2022-04-14 10:32:16

		IP               Opened  Ports
		192.168.2.4          3   135 139 445
		192.168.2.1          3   53 80 443

		=================================================
	*/
	var builder = strings.Builder{}
	now := time.Now()
	builder.WriteString(gs("\n=================================================\n\n"))
	builder.WriteString(bb("Total:  ") + bs("%d hosts  x  %d ports", len(ss.scanConfig.Hosts), len(ss.scanConfig.Ports)) + "\n")
	builder.WriteString(bb("Start:  ") + bs(ss.start.Format("2006-01-02 15:04:05")) + bb("  Cost: ") + bs(now.Sub(ss.start).String()) + "\n")
	builder.WriteString(bb("Finish: ") + bs(now.Format("2006-01-02 15:04:05")) + "\n\n")

	builder.WriteString(cb("IP") + "               " + cb("Opened") + "  " + cb("Ports") + "\n")

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
		builder.WriteString(cs("%-15s", host) + "  " + gb("%5d", len(openedPorts)) + "   ")
		for i, p := range openedPorts {
			if sep := i != 0; sep {
				builder.WriteString(" ")
			}
			builder.WriteString(gb(strconv.Itoa(p)))
		}

		builder.WriteString("\n")
	}
	builder.WriteString(gs("\n=================================================\n"))

	return builder.String()
}
