package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/pets"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Buy(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == "" {
		return List(rest, user, room, flags)
	}

	targetMobInstanceId := 0
	targetUserId := 0

	itemname := rest

	// See if a "from" target was specified: "buy itemname from shopkeepername"
	args := util.SplitButRespectQuotes(strings.ToLower(rest))
	if len(args) >= 3 {
		if args[len(args)-2] == `from` {
			targetUserId, targetMobInstanceId = room.FindByName(args[len(args)-1])

			if user.UserId == targetUserId {
				user.SendText("You can't buy from yourself.")
				return true, nil
			}

			// If nobody found when clearly specified somebody, send an error and abort
			if targetUserId == 0 && targetMobInstanceId == 0 {
				user.SendText("Visit a merchant to buy objects.")
				return true, nil
			}

			itemname = strings.Join(args[0:len(args)-2], ` `) // reform the purchase arg
		}
	}

	success := false
	defer func() {
		mudlog.Debug("PURCHASE", "rest", rest, "itemname", itemname, "targetUserId", targetUserId, "targetMobInstanceId", targetMobInstanceId, "success", success)
	}()

	merchantPlayers := room.GetPlayers(rooms.FindMerchant)
	merchantMobs := room.GetMobs(rooms.FindMerchant)

	for _, uid := range merchantPlayers {
		if targetUserId > 0 && uid != targetUserId {
			continue
		}

		shopUser := users.GetByUserId(uid)
		if shopUser == nil {
			continue
		}

		if success = tryPurchase(itemname, user, room, nil, shopUser); success {
			return true, nil
		}
	}

	for _, miid := range merchantMobs {
		if targetMobInstanceId > 0 && miid != targetMobInstanceId {
			continue
		}

		shopMob := mobs.GetInstance(miid)
		if shopMob == nil {
			continue
		}

		shopMob.Character.Shop.Restock()

		if success = tryPurchase(itemname, user, room, shopMob, nil); success {
			return true, nil
		}
	}

	return true, nil

}

// TODO: This would sure be a lot more straightforward with an interface...
func tryPurchase(request string, user *users.UserRecord, room *rooms.Room, shopMob *mobs.Mob, shopUser *users.UserRecord) bool {

	nameToShopItem := map[string]characters.ShopItem{}

	itemNames := []string{}
	itemNamesFancy := []string{}
	itemPrices := map[int]int{}

	mercNames := []string{}
	mercPrices := map[int]int{}

	buffNames := []string{}
	buffPrices := map[int]int{}

	petNames := []string{}
	petPrices := map[string]int{}

	var saleItems characters.Shop
	if shopMob != nil {
		saleItems = shopMob.Character.Shop.GetInstock()
	} else if shopUser != nil {
		saleItems = shopUser.Character.Shop.GetInstock()
	}

	for _, saleItem := range saleItems {

		if saleItem.ItemId > 0 {
			item := items.New(saleItem.ItemId)
			if item.ItemId == 0 {
				continue
			}
			itemNames = append(itemNames, item.GetSpec().Name)
			itemNamesFancy = append(itemNamesFancy, item.DisplayName())
			nameToShopItem[item.GetSpec().Name] = saleItem

			price := saleItem.Price
			if price == 0 {
				price = item.GetSpec().Value
			} else if price < 0 {
				price = 0
			}
			itemPrices[saleItem.ItemId] = price

			continue
		}

		if saleItem.MobId > 0 {
			mobInfo := mobs.GetMobSpec(mobs.MobId(saleItem.MobId))
			if mobInfo == nil {
				continue
			}
			mercNames = append(mercNames, mobInfo.Character.Name)
			nameToShopItem[mobInfo.Character.Name] = saleItem

			price := saleItem.Price
			if price == 0 {
				price = 250 * mobInfo.Character.Level
			} else if price < 0 {
				price = 0
			}
			mercPrices[saleItem.MobId] = price

			continue
		}

		if saleItem.BuffId > 0 {
			buffInfo := buffs.GetBuffSpec(saleItem.BuffId)
			if buffInfo == nil {
				continue
			}
			buffNames = append(buffNames, buffInfo.Name)
			nameToShopItem[buffInfo.Name] = saleItem

			price := saleItem.Price
			if price == 0 {
				price = 1000
			} else if price < 0 {
				price = 0
			}
			buffPrices[saleItem.BuffId] = price

			continue
		}

		if saleItem.PetType != `` {
			petInfo := pets.GetPetCopy(saleItem.PetType)
			if !petInfo.Exists() {
				continue
			}
			petNames = append(petNames, petInfo.Type)
			nameToShopItem[petInfo.Type] = saleItem

			price := saleItem.Price
			if price == 0 {
				price = 10000
			} else if price < 0 {
				price = 0
			}

			petPrices[saleItem.PetType] = price

			continue
		}

	}

	allNames := []string{}
	allNames = append(allNames, itemNames...)
	allNames = append(allNames, mercNames...)
	allNames = append(allNames, buffNames...)
	allNames = append(allNames, petNames...)

	match, closeMatch := util.FindMatchIn(request, allNames...)
	if match == `` {
		match = closeMatch
	}

	if match == `` {

		if shopMob != nil {
			extraSay := ``

			if len(itemNames) > 0 {
				randSelection := util.Rand(len(itemNames))
				extraSay = fmt.Sprintf(` Any interest in this <ansi fg="itemname">%s</ansi>?`, itemNamesFancy[randSelection])
			} else if len(buffNames) > 0 {
				randSelection := util.Rand(len(buffNames))
				extraSay = fmt.Sprintf(` Maybe you would enjoy this %s enchantment?`, buffNames[randSelection])
			} else if len(mercNames) > 0 {
				randSelection := util.Rand(len(mercNames))
				extraSay = fmt.Sprintf(` <ansi fg="mobname">%s</ansi> is a loyal mercenary, if you're interested.`, mercNames[randSelection])
			} else if len(petNames) > 0 {
				randSelection := util.Rand(len(petNames))
				extraSay = fmt.Sprintf(` <ansi fg="petname">%s</ansi> is a loyal mercenary, if you're interested.`, petNames[randSelection])
			}

			shopMob.Command(`say Sorry, I can't offer that right now.` + extraSay)
		}

		return false
	}

	matchedShopItem := nameToShopItem[match]
	if !matchedShopItem.Available() {
		if shopMob != nil {
			shopMob.Command(`say I don't have that item for sale right now.`)
		} else if shopUser != nil {
			user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> doesn't have that for sale right now.`, shopUser.Character.Name))
		}
		return false
	}

	price := 0
	if matchedShopItem.ItemId > 0 {
		price = itemPrices[matchedShopItem.ItemId]
	} else if matchedShopItem.MobId > 0 {
		price = mercPrices[matchedShopItem.MobId]
	} else if matchedShopItem.BuffId > 0 {
		price = buffPrices[matchedShopItem.BuffId]
	} else if matchedShopItem.PetType != `` {
		price = petPrices[matchedShopItem.PetType]
	}

	if user.Character.Gold < price {
		if shopMob != nil {
			shopMob.Command(`say You don't have enough gold for that.`)
		} else if shopUser != nil {
			user.SendText(`You don't have enough gold for that.`)
		}
		return false
	}

	tradeItemName := ``
	if matchedShopItem.TradeItemId > 0 {
		tradeItm := items.New(matchedShopItem.TradeItemId)
		tradeItemName = tradeItm.Name()
		if _, found := user.Character.FindInBackpack(tradeItemName); !found {
			user.SendText(fmt.Sprintf(`You must have a <ansi fg="itemname">%s</ansi> to trade for that.`, tradeItm.DisplayName()))
			return false
		}
	}

	if matchedShopItem.MobId > 0 {

		maxCharmed := user.Character.GetSkillLevel(skills.Tame) + 1
		if len(user.Character.GetCharmIds()) >= maxCharmed {
			user.SendText(fmt.Sprintf(`You can only have %d mobs following you at a time.`, maxCharmed))
			return false
		}

	}

	if shopMob != nil {

		if !shopMob.Character.Shop.Destock(matchedShopItem) {
			shopMob.Command(`say I don't have that item right now.`)
			return false
		}

	} else if shopUser != nil {
		if !shopUser.Character.Shop.Destock(matchedShopItem) {
			user.SendText(`That's not for sale.`)
			return false
		}
	}

	user.Character.Gold -= price
	if shopMob != nil {
		shopMob.Character.Gold += 1 // only gains 1 gold with each sale
	} else if shopUser != nil {
		shopUser.Character.Gold += price
	}

	tradeInString := ``

	if price > 0 {
		tradeInString = fmt.Sprintf(`<ansi fg="gold">%d gold</ansi>`, price)
	}

	if tradeItemName != `` {
		if itm, found := user.Character.FindInBackpack(tradeItemName); found {
			user.Character.RemoveItem(itm)

			events.AddToQueue(events.ItemOwnership{
				UserId: user.UserId,
				Item:   itm,
				Gained: false,
			})

			if tradeInString != `` {
				tradeInString += fmt.Sprintf(` and a <ansi fg="itemname">%s</ansi>`, itm.DisplayName())
			} else {
				tradeInString = fmt.Sprintf(`a <ansi fg="itemname">%s</ansi>`, itm.DisplayName())
			}
		}
	}

	if tradeInString == `` {
		tradeInString = `nothing`
	}

	if matchedShopItem.ItemId > 0 {
		// Give them the item
		newItm := items.New(matchedShopItem.ItemId)
		user.Character.StoreItem(newItm)
		user.PlaySound(`purchase`, `other`)

		events.AddToQueue(events.ItemOwnership{
			UserId: user.UserId,
			Item:   newItm,
			Gained: true,
		})

		if shopMob != nil {

			user.EventLog.Add(`shop`, fmt.Sprintf(`Purchased a <ansi fg="itemname">%s</ansi> from <ansi fg="mobname">%s</ansi> for %s`, newItm.DisplayName(), shopMob.Character.Name, tradeInString))

			user.SendText(
				fmt.Sprintf(`You buy a <ansi fg="itemname">%s</ansi> from <ansi fg="mobname">%s</ansi> for %s.`, newItm.DisplayName(), shopMob.Character.Name, tradeInString),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> buys a <ansi fg="itemname">%s</ansi> from <ansi fg="mobname">%s</ansi>.`, user.Character.Name, newItm.DisplayName(), shopMob.Character.Name),
				user.UserId,
			)

		} else if shopUser != nil {

			user.EventLog.Add(`shop`, fmt.Sprintf(`Purchased a <ansi fg="itemname">%s</ansi> from <ansi fg="username">%s</ansi> for %s.`, newItm.DisplayName(), shopUser.Character.Name, tradeInString))

			user.SendText(
				fmt.Sprintf(`You buy a <ansi fg="itemname">%s</ansi> from <ansi fg="username">%s</ansi> for %s.`, newItm.DisplayName(), shopUser.Character.Name, tradeInString),
			)

			shopUser.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> purchased the <ansi fg="itemname">%s</ansi> you were selling for %s.`, user.Character.Name, newItm.DisplayName(), tradeInString))

			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> buys a <ansi fg="itemname">%s</ansi> from <ansi fg="mobname">%s</ansi>.`, user.Character.Name, newItm.DisplayName(), shopUser.Character.Name),
				user.UserId, shopUser.UserId)
		}

		return true
	}

	if matchedShopItem.MobId > 0 {
		// Give them the merc

		newMob := mobs.NewMobById(mobs.MobId(matchedShopItem.MobId), user.Character.RoomId)
		// Charm 'em
		newMob.Character.Charm(user.UserId, -2, characters.CharmExpiredRevert)
		user.Character.TrackCharmed(newMob.InstanceId, true)

		room.AddMob(newMob.InstanceId)

		if shopMob != nil {

			user.EventLog.Add(`shop`, fmt.Sprintf(`Hired <ansi fg="mobname">%s</ansi> from <ansi fg="mobname">%s</ansi> for %s.`, newMob.Character.Name, shopMob.Character.Name, tradeInString))

			user.SendText(
				fmt.Sprintf(`You pay %s to <ansi fg="mobname">%s</ansi>.`, tradeInString, shopMob.Character.Name),
			)

			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> pays %s to <ansi fg="mobname">%s</ansi>.`, user.Character.Name, tradeInString, shopMob.Character.Name),
				user.UserId,
			)
		} else if shopUser != nil {

			user.EventLog.Add(`shop`, fmt.Sprintf(`Hired <ansi fg="mobname">%s</ansi> from <ansi fg="username">%s</ansi> for %s.`, newMob.Character.Name, shopUser.Character.Name, tradeInString))

			user.SendText(
				fmt.Sprintf(`You hire <ansi fg="mobname">%s</ansi> from <ansi fg="username">%s</ansi> for %s.`, newMob.Character.Name, shopUser.Character.Name, tradeInString),
			)

			shopUser.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> hired your <ansi fg="mobname">%s</ansi> you were selling for %s.`, user.Character.Name, newMob.Character.Name, tradeInString))

			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> hires a <ansi fg="mobname">%s</ansi> from <ansi fg="username">%s</ansi>.`, user.Character.Name, newMob.Character.Name, shopUser.Character.Name),
				user.UserId, shopUser.UserId)

		}

		newMob.Command(`emote is ready to serve.`)

		return true
	}

	if matchedShopItem.BuffId > 0 {

		buffSpec := buffs.GetBuffSpec(matchedShopItem.BuffId)

		if shopMob != nil {

			user.EventLog.Add(`shop`, fmt.Sprintf(`Purchased a <ansi fg="buff">%s</ansi> enchantment from <ansi fg="mobname">%s</ansi> for %s`, buffSpec.Name, shopMob.Character.Name, tradeInString))

			user.SendText(
				fmt.Sprintf(`You pay %s to <ansi fg="mobname">%s</ansi>.`, tradeInString, shopMob.Character.Name),
			)

			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> pays %s to <ansi fg="mobname">%s</ansi>.`, user.Character.Name, tradeInString, shopMob.Character.Name),
				user.UserId,
			)

		} else if shopUser != nil {

			user.EventLog.Add(`shop`, fmt.Sprintf(`Purchased a <ansi fg="buff">%s</ansi> enchantment from  <ansi fg="username">%s</ansi> for %s`, buffSpec.Name, shopUser.Character.Name, tradeInString))

			user.SendText(
				fmt.Sprintf(`You pay %s to <ansi fg="mobname">%s</ansi>.`, tradeInString, shopUser.Character.Name),
			)

			shopUser.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> pays you %s for an enchantment.`, user.Character.Name, tradeInString))

			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> pays to <ansi fg="username">%s</ansi> for an enchantment.`, user.Character.Name, shopUser.Character.Name),
				user.UserId, shopUser.UserId)

		}

		// Apply the buff
		if shopMob != nil {
			shopMob.Command(`emote mutters a soft incantation.`, 1)
		} else if shopUser != nil {
			shopUser.Command(`emote mutters a soft incantation.`, 1)
		}

		events.AddToQueue(events.Buff{
			UserId:        user.UserId,
			MobInstanceId: 0,
			BuffId:        matchedShopItem.BuffId,
		})

		if shopMob != nil {
			shopMob.Command(`say I've done what I can.`, 1)
		}

		return true
	}

	if matchedShopItem.PetType != `` {

		petInfo := pets.GetPetCopy(matchedShopItem.PetType)

		if shopMob != nil {

			user.EventLog.Add(`shop`, fmt.Sprintf(`Purchased a %s pet from <ansi fg="mobname">%s</ansi> for %s`, petInfo.DisplayName(), shopMob.Character.Name, tradeInString))

			user.SendText(
				fmt.Sprintf(`You pay %s to <ansi fg="mobname">%s</ansi>.`, tradeInString, shopMob.Character.Name),
			)

			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> pays %s to <ansi fg="mobname">%s</ansi>.`, user.Character.Name, tradeInString, shopMob.Character.Name),
				user.UserId,
			)

		} else if shopUser != nil {

			user.EventLog.Add(`shop`, fmt.Sprintf(`Purchased a %s pet from <ansi fg="mobname">%s</ansi> for %s`, petInfo.DisplayName(), shopUser.Character.Name, tradeInString))

			user.SendText(
				fmt.Sprintf(`You pay %s to <ansi fg="mobname">%s</ansi>.`, tradeInString, shopUser.Character.Name),
			)

			shopUser.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> pays you %s for the %s.`, user.Character.Name, tradeInString, petInfo.DisplayName()))

			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> pays to <ansi fg="username">%s</ansi> for the %s.`, user.Character.Name, shopUser.Character.Name, petInfo.DisplayName()),
				user.UserId, shopUser.UserId)

		}

		// Apply the buff
		if shopMob != nil {
			shopMob.Command(fmt.Sprintf(`say Take care of your %s, it will always be loyal to you.`, petInfo.DisplayName()), 1)
			shopMob.Command(`say You can name your pet with the <ansi fg="command">pet</ansi> command.`, 1)
		}

		if user.Character.Pet.Exists() {

			if len(user.Character.Pet.Items) > 0 {

				room.SendText(fmt.Sprintf(`%s drops everything they were carrying.`, user.Character.Pet.DisplayName()))

				for _, item := range user.Character.Pet.Items {
					room.AddItem(item, false)
				}
			}

			room.SendText(fmt.Sprintf(`%s sadly slinks away into the shadows. Never to be seen again.`, user.Character.Pet.DisplayName()))
		}

		for i := 0; i < 5; i++ {
			petInfo.Food.Add()
		}

		petInfo.Name = petInfo.Type
		user.Character.Pet = petInfo
		// make sure new pet buffs get applied
		user.Character.Validate(true)

		return true
	}

	return false
}
