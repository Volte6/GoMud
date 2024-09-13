package usercommands

import (
	"fmt"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Appraise(rest string, userId int) (bool, string, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	for _, mobId := range room.GetMobs(rooms.FindMerchant) {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		if rest == "" {

			mob.Command(`say I will appraise items for 20 gold.`)

			return true, ``, nil
		}

		item, found := user.Character.FindInBackpack(rest)
		if !found {
			user.SendText("You don't have that item.")
			return true, ``, nil
		}

		itemSpec := item.GetSpec()
		if itemSpec.ItemId < 1 {
			return true, ``, nil
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

			return true, ``, nil
		}

		user.Character.Gold -= appraisePrice
		mob.Character.Gold += appraisePrice

		user.SendText(fmt.Sprintf(`You give <ansi fg="mobname">%s</ansi> %d gold to appraise <ansi fg="itemname">%s</ansi>.`, mob.Character.Name, appraisePrice, itemSpec.Name))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> appraises <ansi fg="itemname">%s</ansi>.`, user.Character.Name, itemSpec.Name), userId)

		inspectTxt, _ := templates.Process("descriptions/inspect", details)
		user.SendText(inspectTxt)

		break
	}

	return true, ``, nil
}
