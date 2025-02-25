// Round ticks for players
package hooks

import (
	"log/slog"
	"time"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/leaderboard"
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
		slog.Error("Event", "Expected Type", "NewTurn", "Actual Type", e.Type())
		return false
	}

	c := evt.Config.(configs.Config)

	if evt.TurnNumber%uint64(c.TurnsPerAutoSave()) == 0 {
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

		util.TrackTime(`Save Game State`, time.Since(tStart).Seconds())

		// Do leaderboard updates here too
		events.AddToQueue(events.Broadcast{
			Text: `Updating leaderboards...`,
		})

		tStart = time.Now()

		leaderboard.Update()

		util.TrackTime(`Leaderboards`, time.Since(tStart).Seconds())

		events.AddToQueue(events.Broadcast{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})
	}

	return true
}
