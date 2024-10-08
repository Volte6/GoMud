package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Shout(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)
	isDrunk := user.Character.HasBuffFlag(buffs.Drunk)

	rest = strings.ToUpper(rest)

	if isDrunk {
		// modify the text to look like it's the speech of a drunk person
		rest = drunkify(rest)
	}

	if isSneaking {
		room.SendText(fmt.Sprintf(`someone shouts, "<ansi fg="yellow">%s</ansi>"`, rest), user.UserId)
	} else {
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> shouts, "<ansi fg="yellow">%s</ansi>"`, user.Character.Name, rest), user.UserId)
	}

	for _, roomInfo := range room.Exits {
		if otherRoom := rooms.LoadRoom(roomInfo.RoomId); otherRoom != nil {
			if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
				otherRoom.SendText(fmt.Sprintf(`Someone shouts from the <ansi fg="exit">%s</ansi> direction, "<ansi fg="yellow">%s</ansi>"`, sourceExit, rest), user.UserId)
			}
		}
	}

	user.SendText(fmt.Sprintf(`You shout, "<ansi fg="yellow">%s</ansi>"`, rest))

	return true, nil
}
