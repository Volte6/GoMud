package usercommands

import (
	"github.com/volte6/mud/users"
)

func Save(rest string, user *users.UserRecord) (bool, error) {

	user.SendText("Saving...")
	users.SaveUser(*user)
	user.SendText("done.")

	return true, nil
}
