package usercommands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/colorpatterns"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/pets"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
)

func List(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	listedSomething := false

	for _, mobId := range room.GetMobs(rooms.FindMerchant) {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		/// Run restock routine
		mob.Character.Shop.Restock()

		listedSomething = true

		itemsAvailable := characters.Shop{}
		mercsAvailable := characters.Shop{}
		buffsAvailable := characters.Shop{}
		petsAvailable := characters.Shop{}

		for _, saleItem := range mob.Character.Shop.GetInstock() {

			if saleItem.ItemId > 0 {
				itemsAvailable = append(itemsAvailable, saleItem)
				continue
			}

			if saleItem.MobId > 0 {
				mercsAvailable = append(mercsAvailable, saleItem)
				continue
			}

			if saleItem.BuffId > 0 {
				buffsAvailable = append(buffsAvailable, saleItem)
			}

			if saleItem.PetType != `` {
				petsAvailable = append(petsAvailable, saleItem)
			}

		}

		if len(itemsAvailable) == 0 && len(mercsAvailable) == 0 && len(buffsAvailable) == 0 && len(petsAvailable) == 0 {
			mob.Command(`say I have nothing to sell right now, but check again later.`)
			continue
		}

		if len(itemsAvailable) > 0 {

			headers := []string{"Qty", "Name", "Type", "Price"}
			rows := [][]string{}

			for _, stockItm := range itemsAvailable {
				item := items.New(stockItm.ItemId)

				qtyStr := `N/A`
				if stockItm.QuantityMax != 0 {
					qtyStr = strconv.Itoa(stockItm.Quantity)
				}

				price := stockItm.Price
				if price == 0 {
					price = item.GetSpec().Value
				}

				rows = append(rows, []string{
					qtyStr,
					fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, item.DisplayName()) + strings.Repeat(" ", 30-len(item.Name())),
					string(item.GetSpec().Type),
					strconv.Itoa(price)},
				)
			}

			sort.Slice(rows, func(i, j int) bool {
				num1, _ := strconv.Atoi(rows[i][3])
				num2, _ := strconv.Atoi(rows[j][3])
				return num1 < num2
			})

			onlineTableData := templates.GetTable(fmt.Sprintf(`%s by <ansi fg="mobname">%s</ansi>`, colorpatterns.ApplyColorPattern(`Items for sale`, `cyan`), mob.Character.Name), headers, rows)
			tplTxt, _ := templates.Process("tables/shoplist", onlineTableData)
			user.SendText(tplTxt)
			user.SendText(fmt.Sprintf(`To buy something, type: <ansi fg="command">buy [name]</ansi>%s`, term.CRLFStr))
		}

		if len(mercsAvailable) > 0 {

			headers := []string{"Qty", "Name", "Level", "Race", "Price"}

			rows := [][]string{}

			for _, stockMerc := range mercsAvailable {

				mobInfo := mobs.GetMobSpec(mobs.MobId(stockMerc.MobId))
				if mobInfo == nil {
					continue
				}
				raceInfo := races.GetRace(mobInfo.Character.RaceId)
				if raceInfo == nil {
					continue
				}

				qtyStr := `N/A`
				if stockMerc.QuantityMax != 0 {
					qtyStr = strconv.Itoa(stockMerc.Quantity)
				}

				price := stockMerc.Price
				if price == 0 {
					price = 250 * mobInfo.Character.Level
				}

				rows = append(rows, []string{
					qtyStr,
					`<ansi fg="mobname">` + mobInfo.Character.Name + `</ansi>` + strings.Repeat(" ", 30-len(mobInfo.Character.Name)),
					strconv.Itoa(mobInfo.Character.Level),
					raceInfo.Name,
					strconv.Itoa(price),
				})

			}

			sort.Slice(rows, func(i, j int) bool {
				num1, _ := strconv.Atoi(rows[i][4])
				num2, _ := strconv.Atoi(rows[j][4])
				return num1 < num2
			})

			onlineTableData := templates.GetTable(fmt.Sprintf(`%s by <ansi fg="mobname">%s</ansi>`, colorpatterns.ApplyColorPattern(`Mercenaries for hire`, `flame`), mob.Character.Name), headers, rows)
			tplTxt, _ := templates.Process("tables/shoplist", onlineTableData)
			user.SendText(tplTxt)
			user.SendText(fmt.Sprintf(`To Hire a merc, type: <ansi fg="command">hire [name]</ansi>%s`, term.CRLFStr))
		}

		if len(buffsAvailable) > 0 {

			headers := []string{"Qty", "Name", "Price"}
			rows := [][]string{}

			for _, stockBuff := range buffsAvailable {

				buffInfo := buffs.GetBuffSpec(stockBuff.BuffId)
				if buffInfo == nil {
					continue
				}

				qtyStr := `N/A`
				if stockBuff.QuantityMax != 0 {
					qtyStr = strconv.Itoa(stockBuff.Quantity)
				}

				rows = append(rows, []string{
					qtyStr,
					buffInfo.Name + strings.Repeat(" ", 30-len(buffInfo.Name)),
					strconv.Itoa(stockBuff.Price)},
				)
			}

			sort.Slice(rows, func(i, j int) bool {
				num1, _ := strconv.Atoi(rows[i][2])
				num2, _ := strconv.Atoi(rows[j][2])
				return num1 < num2
			})

			onlineTableData := templates.GetTable(fmt.Sprintf(`%s by <ansi fg="mobname">%s</ansi>`, colorpatterns.ApplyColorPattern(`Enchantments`, `rainbow`), mob.Character.Name), headers, rows)
			tplTxt, _ := templates.Process("tables/shoplist", onlineTableData)
			user.SendText(tplTxt)
			user.SendText(fmt.Sprintf(`To buy an enchantment, type: <ansi fg="command">buy [name]</ansi>%s`, term.CRLFStr))
		}

		if len(petsAvailable) > 0 {

			headers := []string{"Qty", "Type", "Price"}
			rows := [][]string{}

			for _, stockPet := range petsAvailable {

				petInfo := pets.GetPetCopy(stockPet.PetType)
				if !petInfo.Exists() {
					continue
				}

				qtyStr := `N/A`
				if stockPet.QuantityMax != 0 {
					qtyStr = strconv.Itoa(stockPet.Quantity)
				}

				price := stockPet.Price
				if price == 0 {
					price = 10000
				}

				rows = append(rows, []string{
					qtyStr,
					`<ansi fg="petname">` + petInfo.Type + strings.Repeat(" ", 30-len(petInfo.Type)) + `</ansi>`,
					strconv.Itoa(price)},
				)
			}

			sort.Slice(rows, func(i, j int) bool {
				num1, _ := strconv.Atoi(rows[i][2])
				num2, _ := strconv.Atoi(rows[j][2])
				return num1 < num2
			})

			onlineTableData := templates.GetTable(fmt.Sprintf(`%s by <ansi fg="mobname">%s</ansi>`, colorpatterns.ApplyColorPattern(`Pets`, `turquoise`), mob.Character.Name), headers, rows)
			tplTxt, _ := templates.Process("tables/shoplist", onlineTableData)
			user.SendText(tplTxt)
			user.SendText(fmt.Sprintf(`To buy a pet, type: <ansi fg="command">buy [name]</ansi>%s`, term.CRLFStr))
		}
	}

	for _, uid := range room.GetPlayers(rooms.FindMerchant) {

		if uid == user.UserId {
			continue
		}

		shopUser := users.GetByUserId(uid)
		if shopUser == nil {
			continue
		}

		listedSomething = true

		itemsAvailable := characters.Shop{}
		mercsAvailable := characters.Shop{}
		buffsAvailable := characters.Shop{}
		petsAvailable := characters.Shop{}

		for _, saleItem := range shopUser.Character.Shop.GetInstock() {

			if saleItem.ItemId > 0 {
				itemsAvailable = append(itemsAvailable, saleItem)
				continue
			}

			if saleItem.MobId > 0 {
				mercsAvailable = append(mercsAvailable, saleItem)
				continue
			}

			if saleItem.BuffId > 0 {
				buffsAvailable = append(buffsAvailable, saleItem)
			}

			if saleItem.PetType != `` {
				petsAvailable = append(petsAvailable, saleItem)
			}
		}

		if len(itemsAvailable) == 0 && len(mercsAvailable) == 0 && len(buffsAvailable) == 0 && len(petsAvailable) == 0 {
			continue
		}

		if len(itemsAvailable) > 0 {

			headers := []string{"Qty", "Name", "Type", "Price"}
			rows := [][]string{}

			for _, stockItm := range itemsAvailable {
				item := items.New(stockItm.ItemId)

				qtyStr := `N/A`
				if stockItm.QuantityMax != 0 {
					qtyStr = strconv.Itoa(stockItm.Quantity)
				}

				price := stockItm.Price
				if price == 0 {
					price = item.GetSpec().Value
				}

				rows = append(rows, []string{
					qtyStr,
					fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, item.DisplayName()) + strings.Repeat(" ", 30-len(item.Name())),
					string(item.GetSpec().Type),
					strconv.Itoa(price)},
				)
			}

			sort.Slice(rows, func(i, j int) bool {
				num1, _ := strconv.Atoi(rows[i][3])
				num2, _ := strconv.Atoi(rows[j][3])
				return num1 < num2
			})

			onlineTableData := templates.GetTable(fmt.Sprintf(`%s by <ansi fg="username">%s</ansi>`, colorpatterns.ApplyColorPattern(`Items for sale`, `cyan`), shopUser.Character.Name), headers, rows)
			tplTxt, _ := templates.Process("tables/shoplist", onlineTableData)
			user.SendText(tplTxt)
			user.SendText(fmt.Sprintf(`To buy something, type: <ansi fg="command">buy [name]</ansi>%s`, term.CRLFStr))
		}

		if len(mercsAvailable) > 0 {

			headers := []string{"Qty", "Name", "Level", "Race", "Price"}

			rows := [][]string{}

			for _, stockMerc := range mercsAvailable {

				mobInfo := mobs.GetMobSpec(mobs.MobId(stockMerc.MobId))
				if mobInfo == nil {
					continue
				}
				raceInfo := races.GetRace(mobInfo.Character.RaceId)
				if raceInfo == nil {
					continue
				}

				qtyStr := `N/A`
				if stockMerc.QuantityMax != 0 {
					qtyStr = strconv.Itoa(stockMerc.Quantity)
				}

				price := stockMerc.Price
				if price == 0 {
					price = 250 * mobInfo.Character.Level
				}

				rows = append(rows, []string{
					qtyStr,
					`<ansi fg="mobname">` + mobInfo.Character.Name + `</ansi>` + strings.Repeat(" ", 30-len(mobInfo.Character.Name)),
					strconv.Itoa(mobInfo.Character.Level),
					raceInfo.Name,
					strconv.Itoa(price),
				})

			}

			sort.Slice(rows, func(i, j int) bool {
				num1, _ := strconv.Atoi(rows[i][4])
				num2, _ := strconv.Atoi(rows[j][4])
				return num1 < num2
			})

			onlineTableData := templates.GetTable(fmt.Sprintf(`%s by <ansi fg="username">%s</ansi>`, colorpatterns.ApplyColorPattern(`Mercenaries for hire`, `flame`), shopUser.Character.Name), headers, rows)
			tplTxt, _ := templates.Process("tables/shoplist", onlineTableData)
			user.SendText(tplTxt)
			user.SendText(fmt.Sprintf(`To Hire a merc, type: <ansi fg="command">hire [name]</ansi>%s`, term.CRLFStr))
		}

		if len(buffsAvailable) > 0 {

			headers := []string{"Qty", "Name", "Price"}
			rows := [][]string{}

			for _, stockBuff := range buffsAvailable {

				buffInfo := buffs.GetBuffSpec(stockBuff.BuffId)
				if buffInfo == nil {
					continue
				}

				qtyStr := `N/A`
				if stockBuff.QuantityMax != 0 {
					qtyStr = strconv.Itoa(stockBuff.Quantity)
				}

				rows = append(rows, []string{
					qtyStr,
					buffInfo.Name + strings.Repeat(" ", 30-len(buffInfo.Name)),
					strconv.Itoa(stockBuff.Price)},
				)
			}

			sort.Slice(rows, func(i, j int) bool {
				num1, _ := strconv.Atoi(rows[i][2])
				num2, _ := strconv.Atoi(rows[j][2])
				return num1 < num2
			})

			onlineTableData := templates.GetTable(fmt.Sprintf(`%s by <ansi fg="username">%s</ansi>`, colorpatterns.ApplyColorPattern(`Enchantments`, `rainbow`), shopUser.Character.Name), headers, rows)
			tplTxt, _ := templates.Process("tables/shoplist", onlineTableData)
			user.SendText(tplTxt)
			user.SendText(fmt.Sprintf(`To buy an enchantment, type: <ansi fg="command">buy [name]</ansi>%s`, term.CRLFStr))
		}

		if len(petsAvailable) > 0 {

			headers := []string{"Qty", "Type", "Price"}
			rows := [][]string{}

			for _, stockPet := range petsAvailable {

				petInfo := pets.GetPetCopy(stockPet.PetType)
				if !petInfo.Exists() {
					continue
				}

				qtyStr := `N/A`
				if stockPet.QuantityMax != 0 {
					qtyStr = strconv.Itoa(stockPet.Quantity)
				}

				price := stockPet.Price
				if price == 0 {
					price = 10000
				}

				rows = append(rows, []string{
					qtyStr,
					`<ansi fg="petname">` + petInfo.Type + strings.Repeat(" ", 30-len(petInfo.Type)) + `</ansi>`,
					strconv.Itoa(price)},
				)
			}

			sort.Slice(rows, func(i, j int) bool {
				num1, _ := strconv.Atoi(rows[i][2])
				num2, _ := strconv.Atoi(rows[j][2])
				return num1 < num2
			})

			onlineTableData := templates.GetTable(fmt.Sprintf(`%s by <ansi fg="username">%s</ansi>`, colorpatterns.ApplyColorPattern(`Pets`, `turquoise`), user.Character.Name), headers, rows)
			tplTxt, _ := templates.Process("tables/shoplist", onlineTableData)
			user.SendText(tplTxt)
			user.SendText(fmt.Sprintf(`To buy a pet, type: <ansi fg="command">buy [name]</ansi>%s`, term.CRLFStr))
		}

	}

	if !listedSomething {
		user.SendText("Visit a merchant to list and buy objects.")
	}

	return true, nil
}
