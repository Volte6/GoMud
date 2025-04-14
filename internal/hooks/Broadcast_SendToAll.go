package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// Checks whether their level is too high for a guide
func Broadcast_SendToAll(e events.Event) events.ListenerReturn {

	broadcast, typeOk := e.(events.Broadcast)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Broadcast", "Actual Type", e.Type())
		return events.Continue
	}

	textOut := ``
	if len(broadcast.Text) > 0 {
		textOut = templates.AnsiParse(broadcast.Text)
	}

	textOutSR := ``
	if len(broadcast.TextScreenReader) > 0 {
		textOutSR = templates.AnsiParse(broadcast.TextScreenReader)
	}

	for _, u := range users.GetAllActiveUsers() {

		if broadcast.IsCommunication {
			if u.Deafened && !broadcast.SourceIsMod {
				continue
			}
		}

		events.AddToQueue(events.RedrawPrompt{UserId: u.UserId}, 100)

		if u.ScreenReader {

			if len(textOutSR) > 0 {

				if broadcast.SkipLineRefresh {
					connections.SendTo(
						[]byte(textOutSR),
						u.ConnectionId(),
					)
				} else {
					connections.SendTo(
						[]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+textOutSR),
						u.ConnectionId(),
					)
				}

				continue
			}

		}

		if broadcast.SkipLineRefresh {
			connections.SendTo(
				[]byte(textOut),
				u.ConnectionId(),
			)
		} else {
			connections.SendTo(
				[]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+textOut),
				u.ConnectionId(),
			)
		}
	}

	return events.Continue
}
