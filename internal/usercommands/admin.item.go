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

	itemId := items.FindItemByName(rest)

	if itemId < 1 {
		itemId, _ = strconv.Atoi(rest)
	}

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

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`item`, rest)

	itemCreateAnswerName := ``
	itemCreateAnswerDescription := `` // description when looking
	itemCreateAnswerValue := 0        // value override
	itemCreateAnswerType := ``
	itemCreateAnswerSubType := ``
	itemCreateAnswerDamage := ``
	itemCreateAnswerMaxUses := 0
	itemCreateAnswerKeyLockId := ``
	itemCreateAnswerQuestToken := ``

	if isNew {
		user.SendText(``)
		user.SendText(fmt.Sprintf(`Lets get a little info first.%s`, term.CRLFStr))
	}

	//
	// Name Selection
	//
	{
		question := cmdPrompt.Ask(`What will the item be called?`, []string{})
		if !question.Done {
			return true, nil
		}

		itemCreateAnswerName = question.Response
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

		question := cmdPrompt.Ask(`What Type of item will it be?`, []string{})
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
				itemCreateAnswerType = t.Type
			}
		}

		if itemCreateAnswerType == `` {
			question.RejectResponse()

			tplTxt, _ := templates.Process("tables/numbered-listtables/numbered-list", typeOptions)
			user.SendText(tplTxt)

			return true, nil
		}
	}

	//
	// Damage (if weapon)
	//
	if itemCreateAnswerType == `weapon` {
		question := cmdPrompt.Ask(`What damage does this weapon do (Example: 1d4)?`, []string{})
		if !question.Done {
			return true, nil
		}

		d := items.Damage{
			DiceRoll: question.Response,
		}
		d.InitDiceRoll(d.DiceRoll)

		itemCreateAnswerDamage = d.FormatDiceRoll()
	}

	//
	// Target room/exit/container (If key)
	//
	if itemCreateAnswerType == `key` {

		question := cmdPrompt.Ask(`What Room Id will this key be used in?`, []string{}, `_`)
		if !question.Done {
			return true, nil
		}

		if question.Response == `_` {
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		roomId, _ := strconv.Atoi(question.Response)
		if roomId == 0 {
			question.RejectResponse()
			return true, nil
		}

		question = cmdPrompt.Ask(`What exit name or container will this open?`, []string{}, `_`)
		if !question.Done {
			return true, nil
		}

		if question.Response == `_` {
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		itemCreateAnswerKeyLockId = fmt.Sprintf(`%d-%s`, roomId, strings.ToLower(question.Response))
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

		question := cmdPrompt.Ask(`What Subtype of item will it be?`, []string{})
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
				itemCreateAnswerSubType = t.Type
			}
		}

		if itemCreateAnswerType == `` {
			question.RejectResponse()

			tplTxt, _ := templates.Process("tables/numbered-listtables/numbered-list", subTypeOptions)
			user.SendText(tplTxt)

			return true, nil
		}
	}

	//
	// Maximum Uses
	//
	{
		question := cmdPrompt.Ask(`Will this item have a maximum number of uses?`, []string{`y`, `n`}, `n`)
		if !question.Done {
			return true, nil
		}

		if question.Response == `y` {
			question := cmdPrompt.Ask(`How many uses will this item have?`, []string{})
			if !question.Done {
				return true, nil
			}

			itemCreateAnswerMaxUses, _ = strconv.Atoi(question.Response)
		}

	}

	//
	// Description
	//
	{
		question := cmdPrompt.Ask(`Quest token given if acquired (if any):`, []string{}, `_`)
		if !question.Done {
			return true, nil
		}

		if question.Response != `_` {

			if quests.GetQuest(question.Response) == nil {
				user.SendText(`Invalid quest token`)
				question.RejectResponse()
				return true, nil
			}

			itemCreateAnswerQuestToken = question.Response
		}
	}

	//
	// Name Selection
	//
	{
		question := cmdPrompt.Ask(`Gold value override (or zero):`, []string{}, `_`)
		if !question.Done {
			return true, nil
		}

		if question.Response == `_` {
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		itemCreateAnswerValue, _ = strconv.Atoi(question.Response)
	}

	//
	// Description
	//
	{
		question := cmdPrompt.Ask(`Enter a description for the item:`, []string{}, `_`)
		if !question.Done {
			return true, nil
		}

		if question.Response != `_` {
			itemCreateAnswerDescription = question.Response
		}
	}

	//
	// Confirm?
	//
	{
		question := cmdPrompt.Ask(`Does this look correct?`, []string{`y`, `n`}, `n`)
		if !question.Done {

			user.SendText(`  <ansi fg="yellow-bold">Name:</ansi>        <ansi fg="white-bold">` + itemCreateAnswerName + `</ansi>`)
			user.SendText(`  <ansi fg="yellow-bold">Desc:</ansi>        <ansi fg="white-bold">` + itemCreateAnswerDescription + `</ansi>`)
			user.SendText(`  <ansi fg="yellow-bold">Type:</ansi>        <ansi fg="white-bold">` + itemCreateAnswerType + `</ansi>`)
			if itemCreateAnswerType == `key` {
				user.SendText(`  <ansi fg="yellow-bold">KeyId:</ansi>       <ansi fg="white-bold">` + itemCreateAnswerKeyLockId + `</ansi>`)
			}
			if itemCreateAnswerType == `weapon` {
				user.SendText(`  <ansi fg="yellow-bold">Damage:</ansi>      <ansi fg="white-bold">` + itemCreateAnswerDamage + `</ansi>`)
			}
			if itemCreateAnswerMaxUses > 0 {
				user.SendText(`  <ansi fg="yellow-bold">Uses:</ansi>        <ansi fg="white-bold">` + strconv.Itoa(itemCreateAnswerMaxUses) + `</ansi>`)
			}
			if itemCreateAnswerValue > 0 {
				user.SendText(`  <ansi fg="yellow-bold">Value:</ansi>        <ansi fg="white-bold">` + strconv.Itoa(itemCreateAnswerValue) + `</ansi>`)
			}
			if itemCreateAnswerQuestToken != `` {
				user.SendText(`  <ansi fg="yellow-bold">Quest Token:</ansi>    <ansi fg="white-bold">` + itemCreateAnswerQuestToken + `</ansi>`)
			}
			user.SendText(`  <ansi fg="yellow-bold">SubType:</ansi>     <ansi fg="white-bold">` + itemCreateAnswerSubType + `</ansi>`)

			return true, nil
		}

		user.ClearPrompt()

		if question.Response != `y` {
			user.SendText("Aborting...")
			return true, nil
		}
	}

	newItemId, err := items.CreateNewItemFile(
		itemCreateAnswerName,
		itemCreateAnswerDescription,
		itemCreateAnswerValue,
		itemCreateAnswerType,
		itemCreateAnswerSubType,
		itemCreateAnswerDamage,
		itemCreateAnswerMaxUses,
		itemCreateAnswerKeyLockId,
		itemCreateAnswerQuestToken,
	)

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
	user.SendText(`  <ansi fg="black-bold">note: Try <ansi fg="command">item spawn ` + itemCreateAnswerName + `</ansi> to test it.`)

	return true, nil
}
