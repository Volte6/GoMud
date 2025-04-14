package hooks

import (
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/users"
)

//
// Checks for quests on the item
//

func PlaySound(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.MSP)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "MSP", "Actual Type", e.Type())
		return events.Cancel
	}

	if evt.UserId < 1 {
		return events.Continue
	}

	if evt.SoundFile == `` {
		return events.Continue
	}

	if user := users.GetByUserId(evt.UserId); user != nil {

		if evt.SoundType == `MUSIC` {

			if user.LastMusic != evt.SoundFile {

				msg := []byte("!!MUSIC(Off)")
				if connections.IsWebsocket(user.ConnectionId()) {

					connections.SendTo(
						msg,
						user.ConnectionId(),
					)

				} else {

					connections.SendTo(
						term.MspCommand.BytesWithPayload(msg),
						user.ConnectionId(),
					)

				}
			}

			user.LastMusic = evt.SoundFile

			msg := []byte("!!MUSIC(" + evt.SoundFile + " V=" + strconv.Itoa(evt.Volume) + " L=-1 C=1)")

			if connections.IsWebsocket(user.ConnectionId()) {

				connections.SendTo(
					msg,
					user.ConnectionId(),
				)

			} else {

				connections.SendTo(
					term.MspCommand.BytesWithPayload(msg),
					user.ConnectionId(),
				)

			}
		} else {

			msg := []byte("!!SOUND(" + evt.SoundFile + " T=" + evt.Category + " V=" + strconv.Itoa(evt.Volume) + ")")

			if connections.IsWebsocket(user.ConnectionId()) {

				connections.SendTo(
					msg,
					user.ConnectionId(),
				)

			} else {

				connections.SendTo(
					term.MspCommand.BytesWithPayload(msg),
					user.ConnectionId(),
				)

			}

		}

	}

	return events.Continue
}
