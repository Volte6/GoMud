package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Rename(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.rename", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(args[0])
	rest = strings.Join(args[1:], " ")

	if !found {
		user.SendText(fmt.Sprintf("You don't have a %s to rename.", rest))
	} else {
		// Swap the item location
		user.Character.RemoveItem(matchItem)
		oldNameSimple := matchItem.DisplayName()
		oldName := matchItem.DisplayName()
		matchItem.Rename(strings.TrimSpace(rest))
		user.Character.StoreItem(matchItem)

		user.SendText(
			fmt.Sprintf(`You chant softly and wave your hand over the <ansi fg="item">%s</ansi>. Success!`, oldNameSimple),
		)
		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> chants softly and waves their hand over <ansi fg="item">%s</ansi>, causing it to glow briefly.`, user.Character.Name, oldName),
			userId,
		)
	}

	return true, nil
}
