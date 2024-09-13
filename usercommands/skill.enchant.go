package usercommands

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Uncurse(rest string, userId int) (util.MessageQueue, error) {
	return Enchant("uncurse "+rest, userId)
}

func Unenchant(rest string, userId int) (util.MessageQueue, error) {
	return Enchant("remove "+rest, userId)
}

/*
Enchant Skill
Level 1 - Enchant a weapon with a damage bonus.
Level 2 - Enchant equipment with a defensive bonus.
Level 3 - Add a stat bonus to a weapon or equipment in addition to the above.
Level 4 - Remove the enchantment or curse from any object.
*/
func Enchant(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Enchant)

	if skillLevel == 0 {
		response.SendUserMessage(userId, "You don't know how to enchant.")
		response.Handled = true
		return response, fmt.Errorf("you don't know how to enchant")
	}

	if len(rest) == 0 {
		response.SendUserMessage(userId, `Type <ansi fg="command">help enchant</ansi> for more information on the enchant skill.`)
		response.Handled = true
		return response, nil
	}

	removeEnchantment := false
	if strings.HasPrefix(rest, `remove `) {
		rest = strings.TrimPrefix(rest, `remove `)
		removeEnchantment = true
	}

	removeCurse := false
	if strings.HasPrefix(rest, `uncurse `) {
		rest = strings.TrimPrefix(rest, `uncurse `)
		removeCurse = true
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		response.SendUserMessage(userId, `You must be more specific.`)
		response.Handled = true
		return response, nil
	}

	onlyConsider := false
	if args[0] == `chance` {
		onlyConsider = true
		args = args[1:]
		rest = strings.Join(args, ` `)
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		response.SendUserMessage(userId, fmt.Sprintf("You don't have a %s to enchant. Is it still worn, perhaps?", rest))
	} else {

		if (matchItem.GetSpec().Type != items.Weapon && matchItem.GetSpec().Subtype != items.Wearable) || matchItem.GetSpec().Type == items.Holdable {
			response.SendUserMessage(userId, `Enchant only works on weapons and armor.`)
			response.Handled = true
			return response, nil
		}

		if removeCurse || removeEnchantment {
			if skillLevel < 4 {
				response.SendUserMessage(userId, `Your skills are not good enough. Type <ansi fg="command">help enchant</ansi> for more information on the enchant skill.`)
				response.Handled = true
				return response, nil
			}
		}

		if removeCurse {

			if !matchItem.IsCursed() {
				response.SendUserMessage(userId, `That's not cursed.`)
				response.Handled = true
				return response, nil
			}

			user.Character.RemoveItem(matchItem)
			matchItem.Uncurse()
			user.Character.StoreItem(matchItem)

			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> holds out their cursed <ansi fg="itemname">%s</ansi> concentrates. An malevolent ethereal whisp seems to float away.`, user.Character.Name, matchItem.DisplayName()),
			)

			response.SendUserMessage(userId,
				fmt.Sprintf(`You remove the curse from the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
			)

			response.Handled = true
			return response, nil
		}

		if removeEnchantment {

			if skillLevel < 4 {
				response.SendUserMessage(userId, `Type <ansi fg="command">help enchant</ansi> for more information on the enchant skill.`)
				response.Handled = true
				return response, nil
			}

			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> holds out their <ansi fg="itemname">%s</ansi> and slowly waves their hand over it. Runes appear to glow on the surface, which fade as they float away.`, user.Character.Name, matchItem.DisplayName()),
			)

			response.SendUserMessage(userId,
				fmt.Sprintf(`You remove the enchantment from the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
			)

			user.Character.RemoveItem(matchItem)
			matchItem.UnEnchant()
			user.Character.StoreItem(matchItem)

			response.Handled = true
			return response, nil

		}

		/*
			if matchItem.IsEnchanted() {
				response.SendUserMessage(userId, fmt.Sprintf(`The <ansi fg="itemname">%s</ansi> is already enchanted.`, matchItem.DisplayName()))
				response.Handled = true
				return response, nil
			}
		*/

		chanceToDestroy := 50
		chanceToDestroy -= skillLevel * 10                              // 10-40% reduction
		chanceToDestroy += int(matchItem.Enchantments) * 20             // 20% greater chance for each enchantment.
		chanceToDestroy -= user.Character.Stats.Mysticism.ValueAdj >> 2 // 1% less chance for each 4 mysticism points

		if onlyConsider {
			response.SendUserMessage(userId,
				fmt.Sprintf(`Your <ansi fg="itemname">%s</ansi> has been enchanted %d times. There is a %d%% chance it would be destroyed.`, matchItem.DisplayName(), matchItem.Enchantments, chanceToDestroy),
			)
			response.Handled = true
			return response, nil
		}

		if !user.Character.TryCooldown(skills.Enchant.String(), configs.GetConfig().MinutesToRounds(15)) {
			response.SendUserMessage(userId,
				fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Enchant.String())),
			)
			response.Handled = true
			return response, errors.New(`you're doing that too often`)
		}

		damageBonus := 0
		defenseBonus := 0
		statBonus := map[string]int{}
		cursed := false

		// At skill level 1, can enchant only weapons
		if matchItem.GetSpec().Type == items.Weapon {
			damageBonus = int(math.Ceil(math.Sqrt(float64(user.Character.Stats.Mysticism.ValueAdj))))
		}

		// At skill level 2, can enchant weapons and armor
		if skillLevel >= 2 && matchItem.GetSpec().Subtype == items.Wearable {
			defenseBonus = int(math.Ceil(math.Sqrt(float64(user.Character.Stats.Mysticism.ValueAdj))))
		}

		// At skill level 3, can  provide stat bonuses
		if skillLevel >= 3 {
			allStats := []string{`strength`, `speed`, `smarts`, `vitality`, `mysticism`, `perception`}

			bonusCt := skillLevel - 1 //
			for i := 0; i < bonusCt; i++ {
				// select a random stat
				chosenStat := allStats[util.Rand(len(allStats))]
				statBonus[chosenStat] = int(math.Ceil(math.Sqrt(float64(user.Character.Stats.Mysticism.ValueAdj))))
			}
		}

		roll := util.Rand(100)

		util.LogRoll(`Enchant->Cursed`, roll, 25)

		if roll < 25 {
			cursed = true
		}

		user.Character.RemoveItem(matchItem)

		roll = util.Rand(100)

		util.LogRoll(`Enchant->Destroy`, roll, chanceToDestroy)

		if roll < chanceToDestroy {
			response.SendUserMessage(userId, fmt.Sprintf(`The <ansi fg="itemname">%s</ansi> explodes in a shower of sparks!`, matchItem.DisplayName()))
			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> holds out their <ansi fg="itemname">%s</ansi> and slowly waves their hand over it. The <ansi fg="itemname">%s</ansi> explodes in a shower of sparks!`, user.Character.Name, matchItem.DisplayName(), matchItem.DisplayName()), userId)
			response.Handled = true
			return response, nil
		}

		matchItem.Enchant(damageBonus, defenseBonus, statBonus, cursed)
		user.Character.StoreItem(matchItem)

		response.SendRoomMessage(user.Character.RoomId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> holds out their <ansi fg="itemname">%s</ansi> and slowly waves their hand over it. The <ansi fg="itemname">%s</ansi> glows briefly for a moment and then the glow fades away.`, user.Character.Name, matchItem.DisplayName(), matchItem.DisplayName()),
		)

		response.SendUserMessage(userId,
			fmt.Sprintf(`You enchant the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
		)

		if cursed {
			response.SendUserMessage(userId, fmt.Sprintf(`%14s  %s`, ``, `<ansi fg="red-bold">CURSED!</ansi>`))
		}

		if damageBonus > 0 {
			response.SendUserMessage(userId, fmt.Sprintf(`%14s: +%d`, `Damage Bonus`, damageBonus))
		}
		if defenseBonus > 0 {
			response.SendUserMessage(userId, fmt.Sprintf(`%14s: +%d`, `Armor`, defenseBonus))
		}
		if len(statBonus) > 0 {
			for statName, statBonusAmt := range statBonus {
				statName = strings.Title(statName)
				response.SendUserMessage(userId, fmt.Sprintf(`%14s: +%d`, statName, statBonusAmt))
			}
		}

	}

	response.Handled = true
	return response, nil
}
