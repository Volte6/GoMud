package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

// Expected format is:
// givequest 1-start
// or
// givequest 1-start Say has anyone seen my locket?
// The second message will only be executed if the quest is successfully given.
func GiveQuest(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	// Don't bother if no players are present
	if room.PlayerCt() < 1 {
		response.Handled = true
		return response, nil
	}

	parts := strings.SplitN(rest, " ", 2)

	questToken := parts[0]
	targetUser := ``

	if len(parts) > 1 {
		targetUser = parts[1]
	}

	if targetUser != `` {
		if uid, _ := room.FindByName(targetUser); uid > 0 {

			events.AddToQueue(events.Quest{
				UserId:     uid,
				QuestToken: questToken,
			})

		}
	} else {
		for _, pId := range room.GetPlayers() {

			events.AddToQueue(events.Quest{
				UserId:     pId,
				QuestToken: questToken,
			})

		}
	}

	response.Handled = true
	return response, nil
}
