package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Checks whether their level is too high for a guide
func Message_SendMessage(e events.Event) events.ListenerReturn {

	message, typeOk := e.(events.Message)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Message", "Actual Type", e.Type())
		return events.Continue
	}

	//mudlog.Debug("Message{}", "userId", message.UserId, "roomId", message.RoomId, "length", len(message.Text), "IsCommunication", message.IsCommunication)

	if message.UserId > 0 {

		if user := users.GetByUserId(message.UserId); user != nil {

			// If they are deafened, they cannot hear user communications
			if message.IsCommunication && user.Deafened {
				return events.Continue
			}

			textOut := templates.AnsiParse(message.Text)
			if user.ScreenReader {
				textOut = util.StripCharsForScreenReaders(textOut)
			}
			connections.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+textOut), user.ConnectionId())

			events.AddToQueue(events.RedrawPrompt{UserId: user.UserId}, 100)

		}
	}

	if message.RoomId > 0 {

		room := rooms.LoadRoom(message.RoomId)
		if room == nil {
			return events.Continue
		}

		for _, userId := range room.GetPlayers() {
			skip := false

			if message.UserId == userId {
				return events.Continue
			}

			exLen := len(message.ExcludeUserIds)
			if exLen > 0 {
				for _, excludeId := range message.ExcludeUserIds {
					if excludeId == userId {
						skip = true
						break
					}
				}
			}

			if skip {
				return events.Continue
			}

			if user := users.GetByUserId(userId); user != nil {

				// If they are deafened, they cannot hear user communications
				if message.IsCommunication && user.Deafened {
					return events.Continue
				}

				// If this is a quiet message, make sure the player can hear it
				if message.IsQuiet {
					if !user.Character.HasBuffFlag(buffs.SuperHearing) {
						return events.Continue
					}
				}

				textOut := templates.AnsiParse(message.Text)
				if user.ScreenReader {
					textOut = util.StripCharsForScreenReaders(textOut)
				}

				connections.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+textOut), user.ConnectionId())

				events.AddToQueue(events.RedrawPrompt{UserId: user.UserId}, 100)

			}
		}

	}
	return events.Continue

}
