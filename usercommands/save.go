package usercommands

import (
	"fmt"

	"github.com/volte6/mud/users"
)

func Save(rest string, userId int) (bool, string, error) {

	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("user %d not found", userId)
	}

	user.SendText("Saving...")
	users.SaveUser(*user)
	user.SendText("done.")

	return true, ``, nil
}
