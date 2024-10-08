package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
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

	toUser.SendText(fmt.Sprintf(`<ansi fg="white">***</ansi> <ansi fg="black-bold"><ansi fg="username">%s</ansi> whispers, "%s"</ansi> <ansi fg="white">***</ansi>`, user.Character.Name, rest))

	return true, nil
}
