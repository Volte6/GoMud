package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/scripting"
)

//
// Checks for quests on the item
//

func CheckItemQuests(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.ItemOwnership)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "ItemOwnership", "Actual Type", e.Type())
		return events.Cancel
	}

	// Only care about users for this stuff
	if evt.UserId == 0 {
		return events.Continue
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

	return events.Continue
}
