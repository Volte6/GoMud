package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Bury(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		user.SendText("Bury what?")
		return true, nil
	}

	if corpse, corpseFound := room.FindCorpse(rest); corpseFound {

		if room.RemoveCorpse(corpse) {

			corpseColor := `mob-corpse`
			if corpse.UserId > 0 {
				corpseColor = `user-corpse`
			}

			user.SendText(fmt.Sprintf(`You bury the <ansi fg="%s">%s corpse</ansi>.`, corpseColor, corpse.Character.Name))
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> buries the <ansi fg="%s">%s corpse</ansi>.`, user.Character.Name, corpseColor, corpse.Character.Name), user.UserId)
			return true, nil

		}

		return true, nil
	}

	user.SendText(fmt.Sprintf("You don't see a %s around for burying.", rest))

	return true, nil
}
