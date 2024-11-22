package web

import (
	"sync"

	"github.com/volte6/gomud/internal/users"
)

type Stats struct {
	OnlineUsers   []users.OnlineInfo
	TelnetPorts   []int
	WebSocketPort int
}

var (
	statsLock   = sync.RWMutex{}
	serverStats = Stats{
		WebSocketPort: 0,
		OnlineUsers:   []users.OnlineInfo{},
		TelnetPorts:   []int{},
	}
)

// Returns a copy of the server stats
func GetStats() Stats {
	statsLock.RLock()
	defer statsLock.RUnlock()

	return serverStats
}

// Returns a copy of the server stats
func UpdateStats(s Stats) {
	statsLock.RLock()
	defer statsLock.RUnlock()

	serverStats = s
}

func (s *Stats) Reset() {
	s.WebSocketPort = 0
	s.OnlineUsers = []users.OnlineInfo{}
	s.TelnetPorts = []int{}
}
