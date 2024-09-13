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

func SayTo(rest string, mobId int) (bool, string, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	// Don't bother if no players are present
	if room.PlayerCt() < 1 {
		return true, ``, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))
	if len(args) < 2 {
		return true, ``, nil
	}

	playerId, mobId := room.FindByName(args[0])
	if playerId == 0 {
		return true, ``, nil
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

	return true, ``, nil
}
