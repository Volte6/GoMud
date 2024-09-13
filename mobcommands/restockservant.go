package mobcommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/util"
)

func RestockServant(rest string, mobId int) (util.MessageQueue, error) {

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

	// Restock a specific mob?
	parts := strings.Split(rest, "/")
	stockMobId, _ := strconv.Atoi(parts[0])

	maxRestock := 1
	if len(parts) > 2 {
		maxRestock, _ = strconv.Atoi(parts[1])
	}

	price := 999999
	if len(parts) > 1 {
		price, _ = strconv.Atoi(parts[2])
	}

	if mobId == 0 {
		response.Handled = true
		return response, nil
	}

	var restocked = false

	for idx, servantInfo := range mob.ShopServants {
		if servantInfo.MobId == mobs.MobId(stockMobId) {
			if mob.ShopServants[idx].Quantity >= maxRestock {
				break
			}
			mob.ShopServants[idx].Quantity++
			mob.ShopServants[idx].Price = price
			restocked = true
		}
	}

	if !restocked {
		restocked = true
		mob.ShopServants = append(mob.ShopServants, mobs.MobForHire{
			MobId:    mobs.MobId(stockMobId),
			Quantity: 1,
			Price:    price,
		})
	}

	if restocked {
		response.SendRoomMessage(mob.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> presents some help for hire`, mob.Character.Name), true)
	}

	response.Handled = true
	return response, nil
}
