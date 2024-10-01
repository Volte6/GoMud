package usercommands

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Character(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if !room.IsCharacterRoom {
		return false, fmt.Errorf(`not in a IsCharacterRoom`)
	}

	// All possible commands:
	// new - reroll current character (if no alts enabled, or create a new one and store the current one)
	// change - change to another character in storage
	// delete - dlete a character from storage

	/*
		user.Character = characters.New()
		rooms.MoveToRoom(user.UserId, -1)
	*/

	if configs.GetConfig().MaxAltCharacters == 0 {
		user.SendText(`<ansi fg="203">Alt character are disabled on this server.</ansi>`)
		return true, errors.New(`alt characters disabled`)
	}

	if user.Character.Level < 5 {
		user.SendText(`<ansi fg="203">You must reach level 5 with this character to access character alts.</ansi>`)
		return true, errors.New(`level 5 minimum`)
	}

	menuOptions := []string{`new`}

	cmdPrompt, isNew := user.StartPrompt(`character`, rest)

	altNames := []string{}
	nameToAlt := map[string]characters.Character{}

	for _, char := range characters.LoadAlts(user.Username) {
		altNames = append(altNames, char.Name)
		nameToAlt[char.Name] = char
	}

	if isNew {

		if len(altNames) > 0 {
			menuOptions = append(menuOptions, `view`)
			menuOptions = append(menuOptions, `change`)
			menuOptions = append(menuOptions, `delete`)

			if user.Permission == users.PermissionAdmin {
				menuOptions = append(menuOptions, `hire`)
			}
		}

		if len(nameToAlt) > 0 {
			altTblTxt := getAltTable(nameToAlt)
			user.SendText(``)
			user.SendText(altTblTxt)
		}

	}

	menuOptions = append(menuOptions, `quit`)

	question := cmdPrompt.Ask(`What would you like to do?`, menuOptions, `quit`)
	if !question.Done {
		return true, nil
	}

	/////////////////////////
	// Leave menu
	/////////////////////////
	if question.Response == `quit` {
		user.ClearPrompt()
		return true, nil
	}

	/////////////////////////
	// Create a new alt
	/////////////////////////
	if question.Response == `new` {

		if len(altNames) >= int(configs.GetConfig().MaxAltCharacters) {
			user.SendText(`<ansi fg="203">You already have too many alts.</ansi>`)
			user.SendText(`<ansi fg="203">You'll need to delete one to create a new one.</ansi>`)

			question.RejectResponse()
			return true, nil
		}

		question := cmdPrompt.Ask(`Are you SURE? (Your current character will be saved here to change back to later)`, []string{`yes`, `no`}, `no`)
		if !question.Done {
			return true, nil
		}

		if question.Response == `no` {
			user.ClearPrompt()
			return true, nil
		}

		newAlts := []characters.Character{}
		for _, char := range nameToAlt {
			newAlts = append(newAlts, char)
		}
		newAlts = append(newAlts, *user.Character)
		characters.SaveAlts(user.Username, newAlts)

		// Send them back to start with a fresh/empty character
		user.Character = characters.New()
		rooms.MoveToRoom(user.UserId, -1)

	}

	/////////////////////////
	// Delete an existing alt
	/////////////////////////
	if question.Response == `delete` {

		if len(nameToAlt) > 0 {
			altTblTxt := getAltTable(nameToAlt)
			user.SendText(``)
			user.SendText(altTblTxt)
		}

		question := cmdPrompt.Ask(`Enter the name of the character you wish to delete:`, []string{})
		if !question.Done {
			return true, nil
		}

		match, closeMatch := util.FindMatchIn(question.Response, altNames...)
		if match == `` {
			match = closeMatch
		}

		if match != `` {

			delChar := nameToAlt[match]

			question := cmdPrompt.Ask(`<ansi fg="red">Are you SURE you want to delete <ansi fg="username">`+delChar.Name+`</ansi>?</ansi>`, []string{`yes`, `no`}, `no`)
			if !question.Done {
				return true, nil
			}

			if question.Response == `no` {
				user.SendText(`<ansi fg="203">Okay. Aborting.</ansi>`)
				user.ClearPrompt()
				return true, nil
			}

			newAlts := []characters.Character{}
			for _, char := range nameToAlt {
				if char.Name != match {
					newAlts = append(newAlts, char)
				}
			}
			characters.SaveAlts(user.Username, newAlts)

			user.SendText(`<ansi fg="username">` + match + `</ansi> <ansi fg="red">is deleted.</ansi>`)
			user.ClearPrompt()
			return true, nil

		}

		user.SendText(`<ansi fg="203">No character with the name <ansi fg="username">` + question.Response + `</ansi> found.</ansi>`)

		user.ClearPrompt()
		return true, nil

	}

	/////////////////////////
	// Swap characters
	/////////////////////////
	if question.Response == `change` {

		if len(nameToAlt) > 0 {
			altTblTxt := getAltTable(nameToAlt)
			user.SendText(``)
			user.SendText(altTblTxt)
		}

		question := cmdPrompt.Ask(`Enter the name of the character you wish to change to:`, []string{})
		if !question.Done {
			return true, nil
		}

		match, closeMatch := util.FindMatchIn(question.Response, altNames...)
		if match == `` {
			match = closeMatch
		}

		if match != `` {

			char := nameToAlt[match]

			question := cmdPrompt.Ask(`<ansi fg="51">Are you SURE you want to change to <ansi fg="username">`+char.Name+`</ansi>?</ansi>`, []string{`yes`, `no`}, `no`)
			if !question.Done {
				return true, nil
			}

			if question.Response == `no` {
				user.SendText(`<ansi fg="203">Okay. Aborting.</ansi>`)
				user.ClearPrompt()
				return true, nil
			}

			oldName := user.Character.Name

			newAlts := []characters.Character{}
			for _, c := range nameToAlt {
				if c.Name != match {
					newAlts = append(newAlts, c)
				}
			}
			newAlts = append(newAlts, *user.Character)
			characters.SaveAlts(user.Username, newAlts)

			char.Validate()
			user.Character = &char

			user.Character.RoomId = room.RoomId

			users.SaveUser(*user)

			user.SendText(term.CRLFStr + `You dematerialize as <ansi fg="username">` + oldName + `</ansi>. and rematerialize as <ansi fg="username">` + char.Name + `</ansi>!` + term.CRLFStr)
			room.SendText(`<ansi fg="username">`+oldName+`</ansi> vanishes, and <ansi fg="username">`+char.Name+`</ansi> appears in a shower of sparks!`, user.UserId)

			user.ClearPrompt()
			return true, nil

		}

		user.SendText(`<ansi fg="203">No character with the name <ansi fg="username">` + question.Response + `</ansi> found.</ansi>`)

		user.ClearPrompt()
		return true, nil

	}

	/////////////////////////
	// View characters
	/////////////////////////
	if question.Response == `view` {

		if len(nameToAlt) > 0 {
			altTblTxt := getAltTable(nameToAlt)
			user.SendText(``)
			user.SendText(altTblTxt)
		}

		question := cmdPrompt.Ask(`Enter the name of the character you wish to view:`, []string{})
		if !question.Done {
			return true, nil
		}

		match, closeMatch := util.FindMatchIn(question.Response, altNames...)
		if match == `` {
			match = closeMatch
		}

		if match != `` {

			char := nameToAlt[match]

			char.Validate()

			tmpChar := user.Character
			user.Character = &char

			Status(``, user, room)

			user.Character = tmpChar

			m := mobs.NewMobById(59, user.Character.RoomId)
			m.Character = char
			room.AddMob(m.InstanceId)
			m.Character.Charm(user.UserId, -1, `suicide vanish`)

			user.ClearPrompt()
			return true, nil

		}

		user.SendText(`<ansi fg="203">No character with the name <ansi fg="username">` + question.Response + `</ansi> found.</ansi>`)

		user.ClearPrompt()
		return true, nil

	}

	/////////////////////////
	// Spawn a helper clone - experimental
	/////////////////////////
	if question.Response == `hire` {

		if len(nameToAlt) > 0 {
			altTblTxt := getAltTable(nameToAlt)
			user.SendText(``)
			user.SendText(altTblTxt)
		}

		question := cmdPrompt.Ask(`Enter the name of the character you wish to hire:`, []string{})
		if !question.Done {
			return true, nil
		}

		match, closeMatch := util.FindMatchIn(question.Response, altNames...)
		if match == `` {
			match = closeMatch
		}

		if match != `` {

			char := nameToAlt[match]

			char.Validate()

			m := mobs.NewMobById(59, user.Character.RoomId)
			m.Character = char
			room.AddMob(m.InstanceId)
			m.Character.Charm(user.UserId, -1, `suicide vanish`)

			user.SendText(`<ansi fg="username">` + m.Character.Name + `</ansi> appears to help you out!`)
			room.SendText(`<ansi fg="username">`+m.Character.Name+`</ansi> appears to help <ansi fg="username">`+user.Character.Name+`</ansi>!`, user.UserId)

			m.Command(`emote waves sheepishly.`, 2)

			user.ClearPrompt()
			return true, nil

		}

		user.SendText(`<ansi fg="203">No character with the name <ansi fg="username">` + question.Response + `</ansi> found.</ansi>`)

		user.ClearPrompt()
		return true, nil

	}

	return true, nil
}

func getAltTable(nameToAlt map[string]characters.Character) string {

	headers := []string{"Name", "Level", "Race", "Profession", "Alignment"}
	rows := [][]string{}

	for _, char := range nameToAlt {

		allRanks := char.GetAllSkillRanks()
		raceName := `Unknown`
		if raceInfo := races.GetRace(char.RaceId); raceInfo != nil {
			raceName = raceInfo.Name
		}

		rows = append(rows, []string{
			fmt.Sprintf(`<ansi fg="username">%s</ansi>`, char.Name),
			strconv.Itoa(char.Level),
			raceName,
			skills.GetProfession(allRanks),
			fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, char.AlignmentName(), char.AlignmentName()),
		})

	}

	sort.Slice(rows, func(i, j int) bool {
		num1, _ := strconv.Atoi(rows[i][1])
		num2, _ := strconv.Atoi(rows[j][1])
		return num1 < num2
	})

	altTableData := templates.GetTable(fmt.Sprintf(`Your alt characters (%d/%d)`, len(nameToAlt), configs.GetConfig().MaxAltCharacters), headers, rows)
	tplTxt, _ := templates.Process("tables/generic", altTableData)

	return tplTxt
}
