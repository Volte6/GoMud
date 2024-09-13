package mobcommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/util"
)

func Restock(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Nothing to restock...
	if !mob.IsMerchant {
		response.Handled = true
		return response, nil
	}

	if rest == "gold" {
		mob.Character.Gold++
		response.Handled = true
		return response, nil
	}

	// If nothing specified, just restock whatever is already there.
	if rest == "" {
		for itemId, _ := range mob.ShopStock {
			mob.ShopStock[itemId]++
		}
		response.Handled = true
		return response, nil
	}

	// Restock a specific item?
	restockId := 0
	maxRestock := 0

	if strings.Contains(rest, "/") {
		parts := strings.Split(rest, "/")
		restockId, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
		maxRestock, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
	} else {
		restockId, _ = strconv.Atoi(rest)
	}

	if restockId == 0 {
		response.Handled = true
		return response, nil
	}

	var restocked = false
	if _, ok := mob.ShopStock[restockId]; !ok {
		mob.ShopStock[restockId] = 0
	}

	if maxRestock == 0 || mob.ShopStock[restockId] < maxRestock {
		mob.ShopStock[restockId]++
		restocked = true
	}

	if restocked {
		response.SendRoomMessage(mob.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> restocks some wares`, mob.Character.Name), true)
	}

	response.Handled = true
	return response, nil
}
