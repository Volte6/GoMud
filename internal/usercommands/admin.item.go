package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/quests"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/util"

	"github.com/volte6/gomud/internal/users"
)

func Item(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if user.Permission != users.PermissionAdmin {
		user.SendText(`<ansi fg="alert-4">Only admins can use this command</ansi>`)
		return true, nil
	}

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 1 {
		infoOutput, _ := templates.Process("admincommands/help/command.item", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	// mob create
	if args[0] == `create` {
		return item_Create(rest, user, room)
	}

	if args[0] == `spawn` {
		return item_Spawn(strings.TrimSpace(rest[5:]), user, room)
	}

	return true, nil
}

func item_Spawn(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	itemId := items.FindItem(rest)
	if itemId != 0 {

		itm := items.New(itemId)
		if itm.ItemId > 0 {
			room.AddItem(itm, false)

			user.SendText(
				fmt.Sprintf(`You wave your hands around and <ansi fg="item">%s</ansi> appears from thin air and falls to the ground.`, itm.DisplayName()),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> waves their hands around and <ansi fg="item">%s</ansi> appears from thin air and falls to the ground.`, user.Character.Name, itm.DisplayName()),
				user.UserId,
			)

			return true, nil
		}

	}

	user.SendText(
		fmt.Sprintf(`Item <ansi fg="itemname">%s</ansi> not found.`, rest),
	)

	return true, nil
}

func item_Create(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	var newItemSpec = items.ItemSpec{}

	args := util.SplitButRespectQuotes(rest)
	if len(args) > 1 {
		if itemId := items.FindItem(args[1]); itemId > 0 {
			newItemSpec = *(items.GetItemSpec(itemId))
		}
	}

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`item`, rest)

	if isNew {
		user.SendText(``)
		user.SendText(fmt.Sprintf(`Lets get a little info first.%s`, term.CRLFStr))
	}

	//
	// Name Selection
	//
	{

		question := cmdPrompt.Ask(`What will the item be called?`, []string{newItemSpec.Name}, newItemSpec.Name)
		if !question.Done {
			return true, nil
		}

		newItemSpec.Name = question.Response

	}

	//
	// Type Selection
	//
	{
		allTypes := items.ItemTypes()

		typeOptions := []templates.NameDescription{}
		for _, t := range allTypes {
			typeOptions = append(typeOptions, templates.NameDescription{
				Name:        t.Type,
				Description: t.Description,
			})
		}

		question := cmdPrompt.Ask(`What Type of item will it be?`, []string{string(newItemSpec.Type)}, string(newItemSpec.Type))
		if !question.Done {
			tplTxt, _ := templates.Process("tables/numbered-list", typeOptions)
			user.SendText(tplTxt)
			return true, nil
		}

		typeNameSelection := question.Response
		if restNum, err := strconv.Atoi(typeNameSelection); err == nil {
			if restNum > 0 && restNum <= len(typeOptions) {
				typeNameSelection = typeOptions[restNum-1].Name
			}
		}

		for _, t := range allTypes {
			if strings.EqualFold(t.Type, typeNameSelection) {
				newItemSpec.Type = items.ItemType(t.Type)
			}
		}

		if newItemSpec.Type == `` {
			question.RejectResponse()

			tplTxt, _ := templates.Process("tables/numbered-list", typeOptions)
			user.SendText(tplTxt)

			return true, nil
		}
	}

	//
	// Damage (if weapon)
	//
	if newItemSpec.Type == items.Weapon {

		question := cmdPrompt.Ask(`What damage does this weapon do (Example: 1d4)?`, []string{newItemSpec.Damage.DiceRoll}, newItemSpec.Damage.DiceRoll)
		if !question.Done {
			return true, nil
		}

		if question.Response != `` {
			newItemSpec.Damage.InitDiceRoll(question.Response)
			newItemSpec.Damage.DiceRoll = newItemSpec.Damage.FormatDiceRoll()
		}

	}

	//
	// Target room/exit/container (If key)
	//
	if newItemSpec.Type == items.Key {

		roomIdStr := ``
		roomExitStr := ``

		keyParts := strings.Split(newItemSpec.KeyLockId, `-`)
		if len(keyParts) == 2 {
			roomIdStr = keyParts[0]
			roomExitStr = keyParts[1]
		}

		question := cmdPrompt.Ask(`What Room Id will this key be used in?`, []string{roomIdStr}, roomIdStr)
		if !question.Done {
			return true, nil
		}

		if question.Response == `` {
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		roomId, _ := strconv.Atoi(question.Response)
		if roomId == 0 {
			question.RejectResponse()
			return true, nil
		}

		question = cmdPrompt.Ask(`What exit name or container will this open?`, []string{roomExitStr}, roomExitStr)
		if !question.Done {
			return true, nil
		}

		if question.Response == `` {
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		newItemSpec.KeyLockId = fmt.Sprintf(`%d-%s`, roomId, strings.ToLower(question.Response))
	}

	//
	// SubType Selection
	//
	{
		allSubTypes := items.ItemSubtypes()

		subTypeOptions := []templates.NameDescription{}
		for _, t := range allSubTypes {
			subTypeOptions = append(subTypeOptions, templates.NameDescription{
				Name:        t.Type,
				Description: t.Description,
			})
		}

		question := cmdPrompt.Ask(`What Subtype of item will it be?`, []string{string(newItemSpec.Subtype)}, string(newItemSpec.Subtype))
		if !question.Done {
			tplTxt, _ := templates.Process("tables/numbered-list", subTypeOptions)
			user.SendText(tplTxt)
			return true, nil
		}

		if question.Response == `` {
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		typeNameSelection := question.Response
		if restNum, err := strconv.Atoi(typeNameSelection); err == nil {
			if restNum > 0 && restNum <= len(subTypeOptions) {
				typeNameSelection = subTypeOptions[restNum-1].Name
			}
		}

		for _, t := range allSubTypes {
			if strings.EqualFold(t.Type, typeNameSelection) {
				newItemSpec.Subtype = items.ItemSubType(t.Type)
			}
		}

		if newItemSpec.Subtype == `` {
			question.RejectResponse()

			tplTxt, _ := templates.Process("tables/numbered-list", subTypeOptions)
			user.SendText(tplTxt)

			return true, nil
		}
	}

	//
	// Maximum Uses
	//
	{
		defaultYN := `n`
		if newItemSpec.Uses > 0 {
			defaultYN = `y`
		}
		question := cmdPrompt.Ask(`Will this item have a maximum number of uses?`, []string{`y`, `n`}, defaultYN)
		if !question.Done {
			return true, nil
		}

		if question.Response == `y` {
			question := cmdPrompt.Ask(`How many uses will this item have?`, []string{strconv.Itoa(newItemSpec.Uses)}, strconv.Itoa(newItemSpec.Uses))
			if !question.Done {
				return true, nil
			}

			newItemSpec.Uses, _ = strconv.Atoi(question.Response)
		}

	}

	//
	// Description
	//
	{
		qToken := newItemSpec.QuestToken
		qTokenDefault := qToken
		if qToken == `` {
			qToken = `_`
		}

		question := cmdPrompt.Ask(`Quest token given if acquired (if any):`, []string{qTokenDefault}, qToken)
		if !question.Done {
			return true, nil
		}

		if question.Response != `_` {

			if quests.GetQuest(question.Response) == nil {
				user.SendText(`Invalid quest token`)
				question.RejectResponse()
				return true, nil
			}

			newItemSpec.QuestToken = question.Response
		}
	}

	//
	// Name Selection
	//
	{
		question := cmdPrompt.Ask(`Gold value override:`, []string{strconv.Itoa(newItemSpec.Value)}, strconv.Itoa(newItemSpec.Value))
		if !question.Done {
			return true, nil
		}
		newItemSpec.Value, _ = strconv.Atoi(question.Response)
	}

	//
	// Description
	//
	{
		question := cmdPrompt.Ask(`Enter a description for the item:`, []string{newItemSpec.Description}, newItemSpec.Description)
		if !question.Done {
			return true, nil
		}

		newItemSpec.Description = question.Response
	}

	//
	// Confirm?
	//
	{
		question := cmdPrompt.Ask(`Does this look correct?`, []string{`y`, `n`}, `n`)
		if !question.Done {

			user.SendText(`  <ansi fg="yellow-bold">Name:</ansi>        <ansi fg="white-bold">` + newItemSpec.Name + `</ansi>`)
			user.SendText(`  <ansi fg="yellow-bold">Desc:</ansi>        <ansi fg="white-bold">` + newItemSpec.Description + `</ansi>`)
			user.SendText(`  <ansi fg="yellow-bold">Type:</ansi>        <ansi fg="white-bold">` + string(newItemSpec.Type) + `</ansi>`)

			if newItemSpec.Type == items.Key {
				user.SendText(`  <ansi fg="yellow-bold">KeyId:</ansi>       <ansi fg="white-bold">` + newItemSpec.KeyLockId + `</ansi>`)
			}
			if newItemSpec.Type == items.Weapon {
				user.SendText(`  <ansi fg="yellow-bold">Damage:</ansi>      <ansi fg="white-bold">` + newItemSpec.Damage.DiceRoll + `</ansi>`)
			}

			if newItemSpec.Uses > 0 {
				user.SendText(`  <ansi fg="yellow-bold">Uses:</ansi>        <ansi fg="white-bold">` + strconv.Itoa(newItemSpec.Uses) + `</ansi>`)
			}

			if newItemSpec.Value > 0 {
				user.SendText(`  <ansi fg="yellow-bold">Value:</ansi>       <ansi fg="white-bold">` + strconv.Itoa(newItemSpec.Value) + `</ansi>`)
			}

			if newItemSpec.QuestToken != `` {
				user.SendText(`  <ansi fg="yellow-bold">Quest Token:</ansi> <ansi fg="white-bold">` + newItemSpec.QuestToken + `</ansi>`)
			}

			user.SendText(`  <ansi fg="yellow-bold">SubType:</ansi>     <ansi fg="white-bold">` + string(newItemSpec.Subtype) + `</ansi>`)

			return true, nil
		}

		user.ClearPrompt()

		if question.Response != `y` {
			user.SendText("Aborting...")
			return true, nil
		}
	}

	newItemId, err := items.CreateNewItemFile(newItemSpec)

	if err != nil {
		user.SendText("Error: " + err.Error())
		return true, nil
	}

	itemInst := items.GetItemSpec(newItemId)

	user.SendText(``)
	user.SendText(`  <ansi bg="red" fg="white-bold">ITEM CREATED</ansi>`)
	user.SendText(``)
	user.SendText(`  <ansi fg="yellow-bold">File Path:</ansi>   <ansi fg="white-bold">` + itemInst.Filepath() + `</ansi>`)
	user.SendText(``)
	user.SendText(`  <ansi fg="black-bold">note: Try <ansi fg="command">item spawn ` + newItemSpec.Name + `</ansi> to test it.`)

	return true, nil
}
