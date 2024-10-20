package webclient

import "sync"

type Stats struct {
	OnlineNow int
}

var (
	statsLock   = sync.RWMutex{}
	serverStats = Stats{}
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
