package usercommands

import (
	"github.com/volte6/mud/gametime"
	"github.com/volte6/mud/users"
)

func Time(rest string, user *users.UserRecord) (bool, error) {

	gd := gametime.GetDate()

	dayNight := `day`
	if gd.Night {
		dayNight = `night`
	}

	user.SendText(`It is now ` + gd.String() + `. It is ` + dayNight + `.`)

	return true, nil
}
