package hooks

import (
	"log/slog"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/users"
)

//
// Prune all buffs that have expired.
//

func PruneBuffs_Listener(e events.Event) bool {

	//evt := e.(events.NewTurn)

	roomsWithPlayers := rooms.GetRoomsWithPlayers()
	for _, roomId := range roomsWithPlayers {
		// Get rooom
		if room := rooms.LoadRoom(roomId); room != nil {

			// Handle outstanding player buffs
			logOff := false
			for _, uId := range room.GetPlayers(rooms.FindBuffed) {

				user := users.GetByUserId(uId)

				logOff = false
				if buffsToPrune := user.Character.Buffs.Prune(); len(buffsToPrune) > 0 {
					for _, buffInfo := range buffsToPrune {
						scripting.TryBuffScriptEvent(`onEnd`, uId, 0, buffInfo.BuffId)

						if buffInfo.BuffId == 0 { // Log them out // logoff // logout
							if !user.Character.HasAdjective(`zombie`) { // if they are currently a zombie, we don't log them out from this buff being removed
								logOff = true
							}
						}
					}

					user.Character.Validate()

					if logOff {
						slog.Info("MEDITATION LOGOFF")
						events.AddToQueue(events.System{Command: "logoff", Data: uId})
					}
				}

			}
		}
	}

	// Handle outstanding mob buffs
	for _, mobInstanceId := range mobs.GetAllMobInstanceIds() {

		mob := mobs.GetInstance(mobInstanceId)

		if buffsToPrune := mob.Character.Buffs.Prune(); len(buffsToPrune) > 0 {
			for _, buffInfo := range buffsToPrune {
				scripting.TryBuffScriptEvent(`onEnd`, 0, mobInstanceId, buffInfo.BuffId)
			}

			mob.Character.Validate()
		}

	}

	return true

}
