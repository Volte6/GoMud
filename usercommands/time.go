package usercommands

import (
	"fmt"

	"github.com/volte6/mud/gametime"
	"github.com/volte6/mud/users"
)

func Time(rest string, userId int) (bool, error) {

	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	gd := gametime.GetDate()

	dayNight := `day`
	if gd.Night {
		dayNight = `night`
	}

	user.SendText(`It is now ` + gd.String() + `. It is ` + dayNight + `.`)

	return true, nil
}
