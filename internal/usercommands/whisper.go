package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Whisper(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 1 {
		user.SendText("Whisper to who?")
		return true, nil
	}

	whisperName := args[0]
	if len(rest) < len(whisperName)+1 {
		user.SendText("You need to specify a message.")
		return true, nil
	}

	rest = strings.TrimSpace(rest[len(whisperName)+1:])

	toUser := users.GetByCharacterName(whisperName)
	if toUser == nil {
		user.SendText("You can't find anyone by that name.")
		return true, nil
	}

	sourceIsMod := user.Permission == users.PermissionAdmin || user.Permission == users.PermissionMod
	targetIsMod := toUser.Permission == users.PermissionAdmin || toUser.Permission == users.PermissionMod

	if user.Muted && !targetIsMod {
		user.SendText(`You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	// Whisper do allow special communication between mods/admins and users
	if toUser.Deafened && !sourceIsMod {
		user.SendText(`That user is <ansi fg="alert-5">DEAFENED</ansi> and cannot receive communications from other players.`)
		return true, nil
	}

	toUser.SendText(fmt.Sprintf(`<ansi fg="white">***</ansi> <ansi fg="black-bold"><ansi fg="username">%s</ansi> whispers, "%s"</ansi> <ansi fg="white">***</ansi>`, user.Character.Name, rest))

	user.SendText(fmt.Sprintf(`You sent a <ansi fg="command">whisper</ansi> to <ansi fg="username">%s</ansi>`, toUser.Character.Name))

	return true, nil
}
