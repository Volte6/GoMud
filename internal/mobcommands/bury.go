package mobcommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Bury(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		return true, nil
	}

	if corpse, corpseFound := room.FindCorpse(rest); corpseFound {

		if room.RemoveCorpse(corpse) {

			corpseColor := `mob-corpse`
			if corpse.UserId > 0 {
				corpseColor = `user-corpse`
			}

			room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> buries the <ansi fg="%s">%s corpse</ansi>.`, mob.Character.Name, corpseColor, corpse.Character.Name))
			return true, nil

		}

		return true, nil
	}

	return true, nil
}
