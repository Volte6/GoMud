package usercommands

import (
	"fmt"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Appraise(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	for _, mobId := range room.GetMobs(rooms.FindMerchant) {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		if rest == "" {

			mob.Command(`say I will appraise items for 20 gold.`)

			response.Handled = true
			return response, nil
		}

		item, found := user.Character.FindInBackpack(rest)
		if !found {
			response.SendUserMessage(user.UserId, "You don't have that item.")
			response.Handled = true
			return response, nil
		}

		itemSpec := item.GetSpec()
		if itemSpec.ItemId < 1 {
			response.Handled = true
			return response, nil
		}

		type inspectDetails struct {
			InspectLevel int
			Item         *items.Item
			ItemSpec     *items.ItemSpec
		}

		details := inspectDetails{
			InspectLevel: 2,
			Item:         &item,
			ItemSpec:     &itemSpec,
		}

		appraisePrice := 20

		if appraisePrice > user.Character.Gold {

			mob.Command(fmt.Sprintf("say That costs %d gold to appraise, which you don't seem to have.", appraisePrice))

			response.Handled = true
			return response, nil
		}

		user.Character.Gold -= appraisePrice
		mob.Character.Gold += appraisePrice

		response.SendUserMessage(userId, fmt.Sprintf(`You give <ansi fg="mobname">%s</ansi> %d gold to appraise <ansi fg="itemname">%s</ansi>.`, mob.Character.Name, appraisePrice, itemSpec.Name))
		response.SendRoomMessage(room.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> appraises <ansi fg="itemname">%s</ansi>.`, user.Character.Name, itemSpec.Name))

		inspectTxt, _ := templates.Process("descriptions/inspect", details)
		response.SendUserMessage(userId, inspectTxt)

		break
	}

	response.Handled = true
	return response, nil
}
