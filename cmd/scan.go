package main

import (
	"fmt"
	"github.com/wux1an/port-scanner"
	"log"
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

		progress(p)
	}()

	go func() {
		defer wg.Done()

		s.Scan()
	}()

	wg.Wait()
}

func progress(p <-chan *scanner.ScanItem) {
	for item := range p {
		if item == nil {
			continue
		}

		if item.Err == nil {
			log.Printf("open %s:%d\n", item.Host, item.Port)
		} else {
			//log.Printf("close %s:%d\n", item.Host, item.Port)
		}
	}
}
