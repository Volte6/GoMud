package hooks

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/term"
)

//
// Checks for quests on the item
//

func BroadcastNewChar(e events.Event) events.ListenerReturn {

	if evt, typeOk := e.(events.CharacterCreated); typeOk {
		events.AddToQueue(events.Broadcast{
			Text: fmt.Sprintf(`<ansi fg="character-joined"><ansi fg="username">%s</ansi> has entered the realm!`, evt.CharacterName) + term.CRLFStr,
		})
		return events.Continue
	}

	if evt, typeOk := e.(events.CharacterChanged); typeOk {
		events.AddToQueue(events.Broadcast{
			Text: fmt.Sprintf(`<ansi fg="character-joined"><ansi fg="username">%s</ansi> has entered the realm!`, evt.CharacterName) + term.CRLFStr,
		})
		return events.Continue
	}

	mudlog.Error("Event", "Expected Type", "CharacterCreated/CharacterChanged", "Actual Type", e.Type())

	return events.Cancel

}
