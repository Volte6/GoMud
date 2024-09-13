package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Reload(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.reload", nil)
		response.Handled = true
		user.SendText(infoOutput)
		return response, nil
	}

	switch strings.ToLower(rest) {
	case `items`:
		items.LoadDataFiles()
		user.SendText(`Items reloaded.`)
	default:
		user.SendText(`Unknown reload command.`)
	}
	response.Handled = true
	return response, nil
}
