package usercommands

import (
	"errors"
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

// Later this can be used for zombie specific decision making, such as when players intentionally zombify and let "ai" take over
func ZombieAct(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if !user.Character.HasAdjective(`zombie`) {
		return false, errors.New(`not a zombie!`)
	}

	if util.Rand(5) == 0 {
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> moans, groans and sways a bit...`, user.Character.Name), user.UserId)
	}

	return true, nil
}
