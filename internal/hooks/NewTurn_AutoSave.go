// Round ticks for players
package hooks

import (
	"time"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/plugins"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

//
// Autosaves users/rooms every so often
//

func AutoSave(e events.Event) bool {

	evt, typeOk := e.(events.NewTurn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "NewTurn", "Actual Type", e.Type())
		return false
	}

	if evt.TurnNumber%uint64(configs.GetTimingConfig().TurnsPerAutoSave()) == 0 {
		tStart := time.Now()

		events.AddToQueue(events.Broadcast{
			Text: `Saving users...`,
		})

		users.SaveAllUsers(true)

		events.AddToQueue(events.Broadcast{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

		events.AddToQueue(events.Broadcast{
			Text: `Saving rooms...`,
		})

		rooms.SaveAllRooms()

		events.AddToQueue(events.Broadcast{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

		// Save plugin states if applicable
		plugins.Save()

		util.TrackTime(`Save Game State`, time.Since(tStart).Seconds())

		// Do leaderboard updates here too
		events.AddToQueue(events.Broadcast{
			Text: `Updating leaderboards...`,
		})

		tStart = time.Now()

		util.TrackTime(`Leaderboards`, time.Since(tStart).Seconds())

		events.AddToQueue(events.Broadcast{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})
	}

	return true
}
