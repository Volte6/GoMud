package badinputtracker

import "sync"

var (
	lock        = sync.Mutex{}
	badCommands = map[string]map[string]int{}
)

func TrackBadCommand(cmd string, rest string) {

	lock.Lock()
	defer lock.Unlock()

	if _, ok := badCommands[cmd]; !ok {
		badCommands[cmd] = map[string]int{}
	}

	badCommands[cmd][rest] = badCommands[cmd][rest] + 1
}

func GetBadCommands() map[string]int {

	lock.Lock()
	defer lock.Unlock()

	ret := map[string]int{}

	for cmd, other := range badCommands {
		for rest, ct := range other {
			ret[cmd+` `+rest] = ct
		}
	}

	return ret
}

func Clear() {
	lock.Lock()
	defer lock.Unlock()

	badCommands = map[string]map[string]int{}
}
