package usercommands

import (
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Reload(rest string, user *users.UserRecord) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.reload", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	switch strings.ToLower(rest) {
	case `items`:
		items.LoadDataFiles()
		user.SendText(`Items reloaded.`)
	default:
		user.SendText(`Unknown reload command.`)
	}
	return true, nil
}
