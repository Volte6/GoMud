package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Shout(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	if user.Muted {
		user.SendText(`You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)
	isDrunk := user.Character.HasBuffFlag(buffs.Drunk)

	rest = strings.ToUpper(rest)

	if isDrunk {
		// modify the text to look like it's the speech of a drunk person
		rest = drunkify(rest)
	}

	if isSneaking {
		room.SendTextCommunication(fmt.Sprintf(`someone shouts, "<ansi fg="yellow">%s</ansi>"`, rest), user.UserId)
	} else {
		room.SendTextCommunication(fmt.Sprintf(`<ansi fg="username">%s</ansi> shouts, "<ansi fg="yellow">%s</ansi>"`, user.Character.Name, rest), user.UserId)
	}

	for _, roomInfo := range room.Exits {
		if otherRoom := rooms.LoadRoom(roomInfo.RoomId); otherRoom != nil {
			if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
				otherRoom.SendTextCommunication(fmt.Sprintf(`Someone shouts from the <ansi fg="exit">%s</ansi> direction, "<ansi fg="yellow">%s</ansi>"`, sourceExit, rest), user.UserId)
			}
		}
	}

	for _, roomInfo := range room.ExitsTemp {
		if otherRoom := rooms.LoadRoom(roomInfo.RoomId); otherRoom != nil {
			if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
				otherRoom.SendTextCommunication(fmt.Sprintf(`Someone shouts from the <ansi fg="exit">%s</ansi> direction, "<ansi fg="yellow">%s</ansi>"`, sourceExit, rest), user.UserId)
			}
		}
	}

	for mut := range room.ActiveMutators {
		spec := mut.GetSpec()
		if len(spec.Exits) == 0 {
			continue
		}
		for _, exitInfo := range spec.Exits {
			if otherRoom := rooms.LoadRoom(exitInfo.RoomId); otherRoom != nil {
				if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
					otherRoom.SendTextCommunication(fmt.Sprintf(`Someone shouts from the <ansi fg="exit">%s</ansi> direction, "<ansi fg="yellow">%s</ansi>"`, sourceExit, rest), user.UserId)
				}
			}
		}
	}

	user.SendText(fmt.Sprintf(`You shout, "<ansi fg="yellow">%s</ansi>"`, rest))

	return true, nil
}
