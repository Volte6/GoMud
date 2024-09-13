package mobcommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
)

func Restock(rest string, mobId int) (bool, string, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("mob %d not found", mobId)
	}

	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	// Nothing to restock...
	if !mob.IsMerchant {
		return true, ``, nil
	}

	if rest == "gold" {
		mob.Character.Gold++
		return true, ``, nil
	}

	// If nothing specified, just restock whatever is already there.
	if rest == "" {
		for itemId, _ := range mob.ShopStock {
			mob.ShopStock[itemId]++
		}
		return true, ``, nil
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
		return true, ``, nil
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
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> restocks some wares`, mob.Character.Name))
	}

	return true, ``, nil
}
