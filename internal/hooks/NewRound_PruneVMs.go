package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/scripting"
)

//
// Prunes VM's every now and then to free up memory
//

func PruneVMs(e events.Event) bool {

	if e.(events.NewRound).RoundNumber%100 == 0 {
		scripting.PruneVMs()
	}

	return true
}
