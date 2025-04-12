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

func AutoSave(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.NewTurn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "NewTurn", "Actual Type", e.Type())
		return events.Cancel
	}

	if evt.TurnNumber%uint64(configs.GetTimingConfig().TurnsPerAutoSave()) == 0 {

		totalTimeStart := time.Now()
		defer func() {
			util.TrackTime(`AutoSave`, time.Since(totalTimeStart).Seconds())
		}()

		//////////////////////////////////////////
		// SAVE ALL USERS
		//////////////////////////////////////////
		events.AddToQueue(events.Broadcast{Text: `Saving users...`})

		users.SaveAllUsers(true)

		events.AddToQueue(events.Broadcast{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

		//////////////////////////////////////////
		// SAVE ALL ROOMS
		//////////////////////////////////////////
		events.AddToQueue(events.Broadcast{Text: `Saving rooms...`})

		rooms.SaveAllRooms()

		events.AddToQueue(events.Broadcast{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

		//////////////////////////////////////////
		// SAVE ALL PLUGINS
		//////////////////////////////////////////
		events.AddToQueue(events.Broadcast{Text: `Saving other...`})
		// Save plugin states if applicable
		plugins.Save()

		events.AddToQueue(events.Broadcast{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

	}

	return events.Continue
}
