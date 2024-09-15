package mobcommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
)

func RestockServant(rest string, mobId int) (bool, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("mob %d not found", mobId)
	}

	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	// Nothing to restock...
	if !mob.IsMerchant {
		return true, nil
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
		return true, nil
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
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> presents some help for hire`, mob.Character.Name))
	}

	return true, nil
}
