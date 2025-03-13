package main

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func (w *World) logOff(userId int) {

	if user := users.GetByUserId(userId); user != nil {

		user.EventLog.Add(`conn`, `Logged off`)

		users.SaveUser(*user)

		events.AddToQueue(events.PlayerDespawn{
			UserId:        user.UserId,
			RoomId:        user.Character.RoomId,
			Username:      user.Username,
			CharacterName: user.Character.Name,
			TimeOnline:    user.GetOnlineInfo().OnlineTimeStr,
		})

	}

}

// Handle dropped players
func (w *World) HandleDroppedPlayers(droppedPlayers []int) {

	if len(droppedPlayers) == 0 {
		return
	}

	for _, userId := range droppedPlayers {
		if user := users.GetByUserId(userId); user != nil {

			user.SendText(`<ansi fg="red">you drop to the ground!</ansi>`)

			if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> <ansi fg="red">drops to the ground!</ansi>`, user.Character.Name),
					user.UserId)
			}
		}
	}

	return
}
