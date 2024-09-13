package usercommands

import (
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/util"
)

func Reload(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.reload", nil)
		response.Handled = true
		response.SendUserMessage(userId, infoOutput, false)
		return response, nil
	}

	switch strings.ToLower(rest) {
	case `items`:
		items.LoadDataFiles()
		response.SendUserMessage(userId, `Items reloaded.`, true)
	default:
		response.SendUserMessage(userId, `Unknown reload command.`, true)
	}
	response.Handled = true
	return response, nil
}
