package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func SayTo(rest string, mobId int) (util.MessageQueue, error) {

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

	toUser := users.GetByUserId(playerId)

	rest = strings.TrimSpace(rest[len(args[0]):])
	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	if isSneaking {
		toUser.SendText(fmt.Sprintf(`someone says to you, "<ansi fg="yellow">%s</ansi>"`, rest))
	} else {
		toUser.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> says to you, "<ansi fg="yellow">%s</ansi>"`, mob.Character.Name, rest))
		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> says to <ansi fg="username">%s</ansi>, "<ansi fg="yellow">%s</ansi>"`, mob.Character.Name, toUser.Character.Name, rest), mobId)
	}

	response.Handled = true
	return response, nil
}
