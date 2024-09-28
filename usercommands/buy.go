package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Buy(rest string, userId int) (bool, error) {

	if rest == "" {
		return List(rest, userId)
	}

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf(`user %d not found`, userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	didPurchase := false

	for _, mobId := range room.GetMobs(rooms.FindMerchant) {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		/// Run restock routine
		mob.Character.Shop.Restock()

		nameToShopItem := map[string]characters.ShopItem{}

		itemNames := []string{}
		itemNamesFancy := []string{}
		itemPrices := map[int]int{}

		mercNames := []string{}
		mercPrices := map[int]int{}

		buffNames := []string{}
		buffPrices := map[int]int{}

		for _, saleItem := range mob.Character.Shop.GetInstock() {

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
				}
				buffPrices[saleItem.BuffId] = price

				continue
			}

		}

		allNames := []string{}
		allNames = append(allNames, itemNames...)
		allNames = append(allNames, mercNames...)
		allNames = append(allNames, buffNames...)

		match, closeMatch := util.FindMatchIn(rest, allNames...)
		if match == "" {
			match = closeMatch
		}

		if match == "" {

			extraSay := ``

			if len(itemNames) > 0 {
				randSelection := util.Rand(len(itemNames))
				extraSay = fmt.Sprintf(` Any interest in this %s?`, itemNamesFancy[randSelection])
			} else if len(buffNames) > 0 {
				randSelection := util.Rand(len(buffNames))
				extraSay = fmt.Sprintf(` Maybe you would enjoy this %s enchantment?`, buffNames[randSelection])
			} else if len(mercNames) > 0 {
				randSelection := util.Rand(len(mercNames))
				extraSay = fmt.Sprintf(` %s is a loyal mercenary, if you're interested.`, mercNames[randSelection])
			}

			mob.Command(`say Sorry, I can't offer that right now.` + extraSay)

			continue
		}

		matchedShopItem := nameToShopItem[match]

		if matchedShopItem.Quantity == 0 && matchedShopItem.QuantityMax > 0 {
			mob.Command(`say I don't have that item right now.`)
			continue
		}

		price := 0
		if matchedShopItem.ItemId > 0 {
			price = itemPrices[matchedShopItem.ItemId]
		} else if matchedShopItem.MobId > 0 {
			price = mercPrices[matchedShopItem.MobId]
		} else if matchedShopItem.BuffId > 0 {
			price = buffPrices[matchedShopItem.BuffId]
		}

		if user.Character.Gold < price {
			mob.Command(`say You don't have enough gold for that.`)
			continue
		}

		if matchedShopItem.MobId > 0 {
			maxCharmed := user.Character.GetSkillLevel(skills.Tame) + 1
			if len(user.Character.GetCharmIds()) >= maxCharmed {
				user.SendText(fmt.Sprintf(`You can only have %d mobs following you at a time.`, maxCharmed))
				continue
			}
		}

		if !mob.Character.Shop.Destock(matchedShopItem) {
			mob.Command(`say I don't have that item right now.`)
			continue
		}

		user.Character.Gold -= price
		mob.Character.Gold += price >> 2 // They only retain 1/4th

		didPurchase = true
		if matchedShopItem.ItemId > 0 {
			// Give them the item
			newItm := items.New(matchedShopItem.ItemId)
			user.Character.StoreItem(newItm)

			iSpec := newItm.GetSpec()
			if iSpec.QuestToken != `` {

				events.AddToQueue(events.Quest{
					UserId:     user.UserId,
					QuestToken: iSpec.QuestToken,
				})

			}

			user.SendText(
				fmt.Sprintf(`You buy a <ansi fg="itemname">%s</ansi> for <ansi fg="gold">%d</ansi> gold.`, newItm.DisplayName(), price),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> buys a <ansi fg="itemname">%s</ansi> from <ansi fg="mobname">%s</ansi>.`, user.Character.Name, newItm.DisplayName(), mob.Character.Name),
				userId,
			)

			break
		}

		if matchedShopItem.MobId > 0 {
			// Give them the merc

			newMob := mobs.NewMobById(mobs.MobId(matchedShopItem.MobId), user.Character.RoomId)
			// Charm 'em
			newMob.Character.Charm(user.UserId, -2, characters.CharmExpiredRevert)
			user.Character.TrackCharmed(newMob.InstanceId, true)

			room.AddMob(newMob.InstanceId)

			user.SendText(
				fmt.Sprintf(`You pay <ansi fg="gold">%d</ansi> gold to <ansi fg="mobname">%s</ansi>.`, price, mob.Character.Name),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> pays <ansi fg="gold">%d</ansi> gold to <ansi fg="mobname">%s</ansi>.`, user.Character.Name, price, mob.Character.Name),
				userId,
			)

			newMob.Command(`emote is ready to serve.`)

			break
		}

		if matchedShopItem.BuffId > 0 {
			// Apply the buff
			mob.Command(`emote mutters a soft incantation.`)

			events.AddToQueue(events.Buff{
				UserId:        user.UserId,
				MobInstanceId: 0,
				BuffId:        matchedShopItem.BuffId,
			})

			mob.Command(`say I've done what I can.`)

			break
		}

	}

	if !didPurchase {
		user.SendText("Visit a merchant to buy objects.")
	}

	return true, nil
}
