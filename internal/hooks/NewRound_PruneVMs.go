package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/scripting"
)

//
// Prunes VM's every now and then to free up memory
//

func PruneVMs(e events.Event) events.ListenerReturn {

	if e.(events.NewRound).RoundNumber%100 == 0 {
		scripting.PruneVMs()
	}

	return events.Continue
}
