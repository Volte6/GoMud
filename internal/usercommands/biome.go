package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Biome(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	biome, ok := rooms.GetBiome(room.Biome)

	if !ok {
		user.SendText(`No biome information found about this area.`)
		return false, fmt.Errorf(`biome %s not found`, room.Biome)
	}

	biomeTxt, _ := templates.Process("descriptions/biome", biome)
	user.SendText(biomeTxt)

	return true, nil
}
