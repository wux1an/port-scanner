package main

import (
	"fmt"
	"testing"
)

func Test_parsePortRange(t *testing.T) {
	var tests = []string{
		"22", "1-1024", "22,234", "22,234,1024-2048",
		"22,", "0.1",
	}
	var ns = []int{
		1, 1024, 2, 1027,
		-1, -1,
	}

	for i, j := range tests {
		ports, err := parsePortRange(j)
		if err != nil {
			fmt.Printf("[ %s ], error: %v\n", j, err)
		} else if ns[i] != -1 && len(ports) != ns[i] {
			t.Failed()
		}
	}
}
