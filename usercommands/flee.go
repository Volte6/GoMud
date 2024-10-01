package usercommands

import (
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/users"
)

func Flee(rest string, user *users.UserRecord) (bool, error) {

	if user.Character.Aggro == nil {
		user.SendText(`You aren't in combat!`)
	} else {
		user.SendText(`You attempt to flee...`)
		user.Character.Aggro.Type = characters.Flee
	}

	return true, nil
}
