package flags

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
)

func HandleFlags() {
	var portsearch string

	flag.StringVar(&portsearch, "port-search", "", "Search for the first 10 open ports: -port-search=30000-40000")

	flag.Parse()

	if portsearch != `` {
		doPortSearch(portsearch)
		os.Exit(0)
	}
}

func doPortSearch(portRangeStr string) {
	portRange := strings.Split(portRangeStr, `-`)

	if len(portRange) < 2 {
		slog.Error("-port-search", "error", "Invalid port range specified")
		return
	}

	portRangeStart, _ := strconv.Atoi(portRange[0])
	portRangeEnd, _ := strconv.Atoi(portRange[1])

	if portRangeStart == 0 || portRangeEnd == 0 || portRangeStart >= portRangeEnd {
		slog.Error("-port-search", "error", "Invalid port range specified")
		return
	}

	slog.Info("-port-search", "message", fmt.Sprintf("Searching for first 10 available ports between %d and %d", portRangeStart, portRangeEnd))

	foundPorts := 0
	for i := portRangeStart; i < portRangeEnd; i++ {

		if !isPortInUse(i) {
			slog.Info("-port-search", "message", "Found port", "port", i)
			foundPorts++
		}
		if foundPorts >= 10 {
			break
		}
	}

	slog.Info("-port-search", "message", fmt.Sprintf("Found %d available ports", foundPorts))

}

func isPortInUse(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	ln.Close()
	return false
}
