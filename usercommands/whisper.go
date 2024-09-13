package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Whisper(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 1 {
		response.SendUserMessage(userId, "Whisper to who?")
		response.Handled = true
		return response, nil
	}

	whisperName := args[0]
	if len(rest) < len(whisperName)+1 {
		response.SendUserMessage(userId, "You need to specify a message.")
		response.Handled = true
		return response, nil
	}

	rest = strings.TrimSpace(rest[len(whisperName)+1:])

	toUser := users.GetByCharacterName(whisperName)
	if toUser == nil {
		response.SendUserMessage(userId, "You can't find anyone by that name.")
		response.Handled = true
		return response, nil
	}

	response.SendUserMessage(toUser.UserId, fmt.Sprintf(`<ansi fg="white">***</ansi> <ansi fg="black-bold"><ansi fg="username">%s</ansi> whispers, "%s"</ansi> <ansi fg="white">***</ansi>`, user.Character.Name, rest))

	response.Handled = true
	return response, nil
}
