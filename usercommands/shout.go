package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Shout(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)
	isDrunk := user.Character.HasBuffFlag(buffs.Drunk)

	rest = strings.ToUpper(rest)

	if isDrunk {
		// modify the text to look like it's the speech of a drunk person
		rest = drunkify(rest)
	}

	if isSneaking {
		room.SendText(fmt.Sprintf(`someone shouts, "<ansi fg="yellow">%s</ansi>"`, rest), userId)
	} else {
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> shouts, "<ansi fg="yellow">%s</ansi>"`, user.Character.Name, rest), userId)
	}

	for _, roomInfo := range room.Exits {
		if otherRoom := rooms.LoadRoom(roomInfo.RoomId); otherRoom != nil {
			if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
				otherRoom.SendText(fmt.Sprintf(`Someone shouts from the <ansi fg="exit">%s</ansi> direction, "<ansi fg="yellow">%s</ansi>"`, sourceExit, rest), userId)
			}
		}
	}

	user.SendText(fmt.Sprintf(`You shout, "<ansi fg="yellow">%s</ansi>"`, rest))

	response.Handled = true
	return response, nil
}
