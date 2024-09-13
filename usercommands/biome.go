package usercommands

import (
	"fmt"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Biome(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	biome, ok := rooms.GetBiome(room.Biome)

	if !ok {
		response.SendUserMessage(userId, `No biome information found about this area.`, true)
		return response, fmt.Errorf(`biome %d not found`, room.Biome)
	}

	biomeTxt, _ := templates.Process("descriptions/biome", biome)
	response.SendUserMessage(userId, biomeTxt, false)

	response.Handled = true
	return response, nil
}
