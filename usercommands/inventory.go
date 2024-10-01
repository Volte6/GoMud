package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Inventory(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	itemNames := []string{}
	itemNamesFormatted := []string{}

	itemList := []items.Item{}

	typeSearchTerms := map[string]items.ItemType{
		`weapons`:   items.Weapon,
		`offhand`:   items.Offhand,
		`holdable`:  items.Holdable,
		`shields`:   items.Offhand,
		`head`:      items.Head,
		`neck`:      items.Neck,
		`body`:      items.Body,
		`armor`:     items.Body,
		`belts`:     items.Belt,
		`gloves`:    items.Gloves,
		`rings`:     items.Ring,
		`legs`:      items.Legs,
		`pants`:     items.Legs,
		`leggings`:  items.Legs,
		`feet`:      items.Feet,
		`potions`:   items.Potion,
		`food`:      items.Food,
		`drinks`:    items.Drink,
		`scrolls`:   items.Scroll,
		`grenades`:  items.Grenade,
		`keys`:      items.Key,
		`gemstones`: items.Gemstone,
	}

	subtypeSearchTerms := map[string]items.ItemSubType{
		`armor`:        items.Wearable,
		`clothing`:     items.Wearable,
		`clothes`:      items.Wearable,
		`wearable`:     items.Wearable,
		`drinks`:       items.Drinkable,
		`drinkable`:    items.Drinkable,
		`food`:         items.Edible,
		`usable`:       items.Usable,
		`throwable`:    items.Throwable,
		`bloudgeoning`: items.Bludgeoning,
		`cleaving`:     items.Cleaving,
		`stabbing`:     items.Stabbing,
		`slashing`:     items.Slashing,
		`shooting`:     items.Shooting,
		`claws`:        items.Claws,
	}

	for _, item := range user.Character.GetAllBackpackItems() {

		foundMatch := false
		if len(rest) > 0 {

			for term, itemType := range typeSearchTerms {
				if strings.HasPrefix(term, rest) {
					if item.GetSpec().Type == itemType {
						itemList = append(itemList, item)
						foundMatch = true
						break
					}
				}
			}

			if foundMatch {
				continue
			}

			for term, itemSubtype := range subtypeSearchTerms {
				if strings.HasPrefix(term, rest) {
					if item.GetSpec().Subtype == itemSubtype {
						itemList = append(itemList, item)
						foundMatch = true
						break
					}
				}
			}

			if foundMatch {
				continue
			}

			//
			// Did not find match, search item name for a possible match.
			//
			for _, part := range util.BreakIntoParts(item.Name()) {
				if strings.HasPrefix(part, rest) {
					itemList = append(itemList, item)
					break
				}

			}

		} else {
			itemList = append(itemList, item)
		}

	}

	for _, item := range itemList {

		iName := item.Name()
		iNameFormatted := fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, item.DisplayName())

		iSpec := item.GetSpec()
		if iSpec.Subtype == items.Drinkable || iSpec.Subtype == items.Edible || iSpec.Subtype == items.Usable || iSpec.Type == items.Lockpicks {
			if iSpec.Uses > 0 { // Does the spec indicate a number of uses?
				iName = fmt.Sprintf(`%s (%d)`, iName, item.Uses)                                               // Display uses left
				iNameFormatted = fmt.Sprintf(`%s <ansi fg="uses-left">(%d)</ansi>`, iNameFormatted, item.Uses) // Display uses left
			}
		}
		itemNames = append(itemNames, iName)
		itemNamesFormatted = append(itemNamesFormatted, iNameFormatted)
	}

	raceInfo := races.GetRace(user.Character.RaceId)

	diceRoll := raceInfo.Damage.DiceRoll
	if user.Character.Equipment.Weapon.ItemId != 0 {
		iSpec := user.Character.Equipment.Weapon.GetSpec()
		diceRoll = iSpec.Damage.DiceRoll
	}

	invData := map[string]any{
		`Equipment`:          &user.Character.Equipment,
		`ItemNames`:          itemNames,
		`ItemNamesFormatted`: itemNamesFormatted,
		`AttackDamage`:       diceRoll,
		`RaceInfo`:           raceInfo,
		`Searching`:          len(rest) > 0,
		`Count`:              fmt.Sprintf(`(%d/%d)`, len(itemList), user.Character.CarryCapacity()),
	}

	tplTxt, _ := templates.Process("character/inventory", invData)
	user.SendText(tplTxt)

	return true, nil
}
