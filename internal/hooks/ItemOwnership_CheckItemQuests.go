package hooks

import (
	"log/slog"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/scripting"
)

//
// Checks for quests on the item
//

func CheckItemQuests(e events.Event) bool {

	evt, typeOk := e.(events.ItemOwnership)
	if !typeOk {
		slog.Error("Event", "Expected Type", "ItemOwnership", "Actual Type", e.Type())
		return false
	}

	// Only care about users for this stuff
	if evt.UserId == 0 {
		return true
	}

	itm := evt.Item.(items.Item)

	if evt.Gained {

		iSpec := itm.GetSpec()
		if iSpec.QuestToken != `` {
			events.AddToQueue(events.Quest{
				UserId:     evt.UserId,
				QuestToken: iSpec.QuestToken,
			})
		}

		scripting.TryItemScriptEvent(`onFound`, itm, evt.UserId)

	} else {

		scripting.TryItemScriptEvent(`onLost`, itm, evt.UserId)

	}

	return true
}
