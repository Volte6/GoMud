package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func SayTo(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	args := util.SplitButRespectQuotes(strings.ToLower(rest))
	if len(args) < 2 {
		response.Handled = true
		return response, nil
	}

	playerId, mobId := room.FindByName(args[0])
	if playerId == 0 {
		response.Handled = true
		return response, nil
	}

	rest = strings.TrimSpace(rest[len(args[0]):])
	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	if isSneaking {
		response.SendUserMessage(playerId, fmt.Sprintf(`someone says to you, "<ansi fg="yellow">%s</ansi>"`, rest), true)
	} else {
		response.SendUserMessage(playerId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> says to you, "<ansi fg="yellow">%s</ansi>"`, mob.Character.Name, rest), true)
	}

	response.Handled = true
	return response, nil
}
