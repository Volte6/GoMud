package usercommands

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Locate(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.locate", nil)
		response.Handled = true
		response.SendUserMessage(userId, infoOutput, false)
		return response, nil
	}

	locateUser := users.GetByCharacterName(rest)
	if locateUser != nil {

		room := rooms.LoadRoom(locateUser.Character.RoomId)
		if room == nil {
			return response, fmt.Errorf(`room %d not found`, locateUser.Character.RoomId)
		}

		response.SendUserMessage(userId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> is in room #<ansi fg="yellow-bold">%d</ansi> - <ansi fg="magenta">%s</ansi> <ansi fg="red">【%s】</ansi>`, locateUser.Character.Name, room.RoomId, room.Title, locateUser.Character.Zone),
			true)

		response.SendUserMessage(locateUser.UserId,
			`You get the feeling someone is looking for you...`,
			true)

	}

	allMobNames := []string{}
	allMobsByName := map[string][]mobs.Mob{}
	allMobCt := 0

	// Now look for mobs?
	headers := []string{"MobName", "Room", "Room Title", "Zone", "Stray"}
	rows := [][]string{}

	listAll := false
	startsWith := false
	endsWith := false
	contains := false

	searchTerm := strings.ToLower(rest)
	if searchTerm == `*` {
		listAll = true
	}
	if strings.HasPrefix(searchTerm, `*`) {
		searchTerm = strings.TrimPrefix(searchTerm, `*`)
		endsWith = true
	}
	if strings.HasSuffix(searchTerm, `*`) {
		searchTerm = strings.TrimSuffix(searchTerm, `*`)
		startsWith = true
	}
	if startsWith && endsWith {
		startsWith = false
		endsWith = false
		contains = true
	}

	for _, mobId := range mobs.GetAllMobInstanceIds() {

		mob := mobs.GetInstance(mobId)

		if !listAll {
			testName := strings.ToLower(mob.Character.Name)
			if contains {
				if !strings.Contains(testName, searchTerm) {
					continue
				}
			} else if startsWith {
				if !strings.HasPrefix(testName, searchTerm) {
					continue
				}
			} else if endsWith {
				if !strings.HasSuffix(testName, searchTerm) {
					continue
				}
			} else if testName != searchTerm {
				continue
			}
		}

		if _, ok := allMobsByName[mob.Character.Name]; !ok {
			allMobsByName[mob.Character.Name] = []mobs.Mob{}
			allMobNames = append(allMobNames, mob.Character.Name)
		}
		allMobsByName[mob.Character.Name] = append(allMobsByName[mob.Character.Name], *mob)
		allMobCt++
	}

	if allMobCt > 0 {

		matchesPerPage := 20
		pageCt := int(math.Ceil(float64(allMobCt) / float64(matchesPerPage)))
		pageNow := 0
		sort.Strings(allMobNames)

		ct := 0
		for _, mobName := range allMobNames {
			for _, mob := range allMobsByName[mobName] {
				room := rooms.LoadRoom(mob.Character.RoomId)

				ct++

				// trunacte room.Title to only 20 chars
				roomTitle := room.Title
				if len(roomTitle) > 24 {
					roomTitle = roomTitle[0:23] + `…`
				}

				mobName := mob.Character.Name
				if mob.Character.Aggro != nil {
					mobName = `*` + mobName
				}
				if len(mobName) > 24 {
					mobName = mobName[0:23] + `…`
				}

				rows = append(rows, []string{
					fmt.Sprintf(`%-24s`, mobName),
					fmt.Sprintf(`%-4d`, mob.Character.RoomId),
					fmt.Sprintf(`%-24s`, roomTitle),
					fmt.Sprintf(`%-14s`, mob.Character.Zone),
					fmt.Sprintf(`%-5s`, fmt.Sprintf(`%d/%d`, len(mob.RoomStack), mob.MaxWander)),
				})

				if ct >= matchesPerPage {
					onlineTableData := templates.GetTable(fmt.Sprintf(`Matches for "%s" [Page %d/%d]`, rest, pageNow+1, pageCt), headers, rows)
					tplTxt, _ := templates.Process("tables/generic", onlineTableData)
					response.SendUserMessage(userId, tplTxt, true)
					rows = [][]string{}
					ct = 0
					pageNow++
					continue
				}
			}
		}

		// Final flush
		if pageNow < pageCt {
			onlineTableData := templates.GetTable(fmt.Sprintf(`Matches for "%s" [Page %d/%d]`, rest, pageNow+1, pageCt), headers, rows)
			tplTxt, _ := templates.Process("tables/generic", onlineTableData)
			response.SendUserMessage(userId, tplTxt, true)
		}

		response.Handled = true
		return response, nil
	}

	response.SendUserMessage(userId,
		fmt.Sprintf("No user or mob found with the name %s", rest),
		true)

	response.Handled = true
	return response, nil
}
