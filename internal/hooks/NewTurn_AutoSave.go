// Round ticks for players
package hooks

import (
	"time"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
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
