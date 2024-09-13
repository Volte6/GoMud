package usercommands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
)

func List(rest string, userId int) (bool, string, error) {

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

	handled := false
	for _, mobId := range room.GetMobs(rooms.FindMerchant) {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		listedSomething := false

		if len(mob.ShopStock) > 0 {

			listedSomething = true

			headers := []string{"Qty", "Name", "Type", "Price"}

			rows := [][]string{}

			if len(mob.ShopStock) < 1 {
				rows = append(rows, []string{"-", "-", "-", "-"})
			} else {
				for itemId, itemQty := range mob.ShopStock {
					item := items.New(itemId)
					if item.ItemId < 1 {

						mob.Command(fmt.Sprintf("say Please alert an admin that item %d is missing from the database.", itemId))

						continue
					}
					rows = append(rows, []string{strconv.Itoa(itemQty),
						item.Name() + strings.Repeat(" ", 30-len(item.Name())),
						string(item.GetSpec().Type),
						strconv.Itoa(item.GetSpec().Value)})
				}
			}

			sort.Slice(rows, func(i, j int) bool {
				num1, _ := strconv.Atoi(rows[i][3])
				num2, _ := strconv.Atoi(rows[j][3])
				return num1 < num2
			})

			onlineTableData := templates.GetTable(fmt.Sprintf(`For Sale by %s`, mob.Character.Name), headers, rows)
			tplTxt, _ := templates.Process("tables/shoplist", onlineTableData)
			user.SendText(tplTxt)
			user.SendText(fmt.Sprintf(`To buy something, type: <ansi fg="command">buy [name]</ansi>%s`, term.CRLFStr))

		}

		if len(mob.ShopServants) > 0 {

			listedSomething = true

			headers := []string{"Quantity", "Name", "Level", "Race", "Price"}

			rows := [][]string{}

			for _, hireInfo := range mob.ShopServants {
				if mobInfo := mobs.GetMobSpec(hireInfo.MobId); mobInfo != nil {
					raceInfo := races.GetRace(mobInfo.Character.RaceId)
					rows = append(rows, []string{
						strconv.Itoa(hireInfo.Quantity),
						mobInfo.Character.Name + strings.Repeat(" ", 30-len(mobInfo.Character.Name)),
						strconv.Itoa(mobInfo.Character.Level),
						raceInfo.Name,
						strconv.Itoa(hireInfo.Price),
					})
				}
			}

			sort.Slice(rows, func(i, j int) bool {
				num1, _ := strconv.Atoi(rows[i][4])
				num2, _ := strconv.Atoi(rows[j][4])
				return num1 < num2
			})

			onlineTableData := templates.GetTable(`Mercs for Hire`, headers, rows)
			tplTxt, _ := templates.Process("tables/shoplist", onlineTableData)
			user.SendText(tplTxt)

			user.SendText(fmt.Sprintf(`To hire a mercenary, type: <ansi fg="command">hire [name]</ansi>%s`, term.CRLFStr))

		}

		if !listedSomething {

			mob.Command(`say I have nothing to sell right  now, but check again later.`)

		}

		handled = true
	}

	if !handled {
		user.SendText("Visit a merchant to list and buy objects.")
	}

	return true, ``, nil
}
