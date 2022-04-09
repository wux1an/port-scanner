package scanner

import (
	"fmt"
	"github.com/malfunkt/iprange"
	"testing"
)

func TestIpRange(_ *testing.T) {
	var tests = []string{
		"192.168.*.*",
		"192.168.1.2/24",
		"github.com", // error
		"192.168.1.2/24,192.168.2.2/24",
	}

	for _, t := range tests {
		var list, err = iprange.ParseList(t)
		if err == nil {
			fmt.Printf("[ %s ], total: %d\n", t, len(list.Expand()))
		} else {
			fmt.Printf("[ %s ], error: %v\n", t, err)
		}
	}
}
