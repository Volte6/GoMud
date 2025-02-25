package hooks

import (
	"log/slog"
	"strconv"

	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

//
// Checks for quests on the item
//

func PlaySound_Listener(e events.Event) bool {

	evt, typeOk := e.(events.MSP)
	if !typeOk {
		slog.Error("Event", "Expected Type", "MSP", "Actual Type", e.Type())
		return false
	}

	if evt.UserId < 1 {
		return true
	}

	if evt.SoundFile == `` {
		return true
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

	return true
}
