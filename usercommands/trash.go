package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Trash(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		response.SendUserMessage(userId, fmt.Sprintf(`You don't have a "%s" to trash.`, rest))
	} else {

		isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

		user.Character.RemoveItem(matchItem)

		response.SendUserMessage(userId,
			fmt.Sprintf(`You trash the <ansi fg="item">%s</ansi> for good.`, matchItem.DisplayName()))

		if !isSneaking {
			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> destroys <ansi fg="item">%s</ansi>...`, user.Character.Name, matchItem.DisplayName()))
		}

		iSpec := matchItem.GetSpec()

		grantXP, xpScale := user.Character.GrantXP(int(float64(iSpec.Value) / 10))

		xpMsgExtra := ``
		if xpScale != 100 {
			xpMsgExtra = fmt.Sprintf(` <ansi fg="yellow">(%d%% scale)</ansi>`, xpScale)
		}

		response.SendUserMessage(user.UserId,
			fmt.Sprintf(`You gained <ansi fg="yellow-bold">%d experience points</ansi>%s!`, grantXP, xpMsgExtra))

		// Trigger lost event
		if scriptResponse, err := scripting.TryItemScriptEvent(`onLost`, matchItem, userId); err == nil {
			response.AbsorbMessages(scriptResponse)
		}

	}

	response.Handled = true
	return response, nil
}
