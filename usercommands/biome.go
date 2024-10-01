package usercommands

import (
	"fmt"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Biome(rest string, user *users.UserRecord) (bool, error) {

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	biome, ok := rooms.GetBiome(room.Biome)

	if !ok {
		user.SendText(`No biome information found about this area.`)
		return false, fmt.Errorf(`biome %d not found`, room.Biome)
	}

	biomeTxt, _ := templates.Process("descriptions/biome", biome)
	user.SendText(biomeTxt)

	return true, nil
}
