package usercommands

import (
	"fmt"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Inventory(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	itemNames := []string{}
	itemNamesFormatted := []string{}

	for _, item := range user.Character.Items {

		iName := item.Name()
		iNameFormatted := fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, iName)

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
	}

	tplTxt, _ := templates.Process("character/inventory", invData)
	response.SendUserMessage(userId, tplTxt, false)

	response.Handled = true
	return response, nil
}
