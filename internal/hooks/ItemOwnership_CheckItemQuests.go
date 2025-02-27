package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/scripting"
)

//
// Checks for quests on the item
//

func CheckItemQuests(e events.Event) bool {

	evt, typeOk := e.(events.ItemOwnership)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "ItemOwnership", "Actual Type", e.Type())
		return false
	}

	// Only care about users for this stuff
	if evt.UserId == 0 {
		return true
	}

	if evt.Gained {

		iSpec := evt.Item.GetSpec()
		if iSpec.QuestToken != `` {
			events.AddToQueue(events.Quest{
				UserId:     evt.UserId,
				QuestToken: iSpec.QuestToken,
			})
		}

		scripting.TryItemScriptEvent(`onFound`, evt.Item, evt.UserId)

	} else {

		scripting.TryItemScriptEvent(`onLost`, evt.Item, evt.UserId)

	}

	return true
}
