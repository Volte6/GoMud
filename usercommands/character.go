package usercommands

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
)

func Character(rest string, userId int) (bool, error) {

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

	if !room.IsCharacterRoom {
		return false, fmt.Errorf(`not in a IsCharacterRoom`)
	}

	// All possible commands:
	// new - reroll current character (if no alts enabled, or create a new one and store the current one)
	// change - change to another character in storage
	// delete - dlete a character from storage

	/*
		user.Character = characters.New()
		rooms.MoveToRoom(userId, -1)
	*/

	config := configs.GetConfig()

	if config.MaxAltCharacters == 0 {
		user.SendText(`Alt character are disabled on this server.`)
		return true, errors.New(`alt characters disabled`)
	}

	if user.Character.Level < 5 {
		user.SendText(`You must reach level 5 with this character to access character alts.`)
		return true, errors.New(`level 5 minimum`)
	}

	menuOptions := []string{`new`}

	cmdPrompt, isNew := user.StartPrompt(`character`, rest)

	alts := characters.LoadAlts(user.Username)

	if isNew {

		if len(alts) > 0 {
			menuOptions = append(menuOptions, `view`)
			menuOptions = append(menuOptions, `change`)
			menuOptions = append(menuOptions, `delete`)

			if user.Permission == users.PermissionAdmin {
				menuOptions = append(menuOptions, `hire`)
			}
		}

		if len(alts) > 0 {

			headers := []string{"Name", "Level", "Race", "Profession", "Alignment"}
			rows := [][]string{}

			for _, char := range alts {

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

			characters.SaveAlts(user.Username, alts)
			altTableData := templates.GetTable(fmt.Sprintf(`Your alt characters (%d/%d)`, len(alts), config.MaxAltCharacters), headers, rows)
			tplTxt, _ := templates.Process("tables/generic", altTableData)
			user.SendText(``)
			user.SendText(tplTxt)
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

		if len(alts) >= int(config.MaxAltCharacters) {
			user.SendText(`You already have too many alts.`)
			user.SendText(`You'll need to delete one to create a new one.`)

			question.RejectResponse()
			return true, nil
		}

		question := cmdPrompt.Ask(`Are you SURE? Your current character will be saved here to change back to later.`, []string{`yes`, `no`}, `no`)
		if !question.Done {
			return true, nil
		}

		if question.Response == `no` {
			user.ClearPrompt()
			return true, nil
		}

		alts = append(alts, *user.Character)
		characters.SaveAlts(user.Username, alts)

		// Send them back to start with a fresh/empty character
		user.Character = characters.New()
		rooms.MoveToRoom(userId, -1)

	}

	/////////////////////////
	// Delete an existing alt
	/////////////////////////
	if question.Response == `delete` {

		question := cmdPrompt.Ask(`Enter the name of the character you wish to delete.`, []string{})

		for i, char := range alts {
			if strings.EqualFold(char.Name, question.Response) {

				question := cmdPrompt.Ask(`<ansi fg="red">Are you SURE you want to delete <ansi fg="username">`+char.Name+`</ansi>?</ansi>`, []string{`yes`, `no`}, `no`)
				if !question.Done {
					return true, nil
				}

				if question.Response == `no` {
					user.SendText(`Okay. Aborting.`)
					user.ClearPrompt()
					return true, nil
				}

				alts = append(alts[:i], alts[i+1:]...)

				characters.SaveAlts(user.Username, alts)

				user.SendText(`<ansi fg="username">` + char.Name + `</ansi> <ansi fg="red">is deleted.</ansi>`)
				user.ClearPrompt()
				return true, nil

			}
		}

		user.SendText(`No character with that name found.`)

		user.ClearPrompt()
		return true, nil

	}

	/////////////////////////
	// Swap characters
	/////////////////////////
	if question.Response == `change` {

		question := cmdPrompt.Ask(`Enter the name of the character you wish to change to.`, []string{})
		if !question.Done {
			return true, nil
		}

		for i, char := range alts {
			if strings.EqualFold(char.Name, question.Response) {

				question := cmdPrompt.Ask(`<ansi fg="51">Are you SURE you want to change to <ansi fg="username">`+char.Name+`</ansi>?</ansi>`, []string{`yes`, `no`}, `no`)
				if !question.Done {
					return true, nil
				}

				if question.Response == `no` {
					user.SendText(`Okay. Aborting.`)
					user.ClearPrompt()
					return true, nil
				}

				oldName := user.Character.Name

				alts = append(alts[:i], alts[i+1:]...)
				alts = append(alts, *user.Character)
				user.Character = &char

				user.Character.Validate()

				characters.SaveAlts(user.Username, alts)
				users.SaveUser(*user)

				user.SendText(term.CRLFStr + `You dematerialize as <ansi fg="username">` + oldName + `</ansi>. and rematerialize as <ansi fg="username">` + char.Name + `</ansi>!` + term.CRLFStr)
				room.SendText(`<ansi fg="username">`+oldName+`</ansi> vanishes, and <ansi fg="username">`+char.Name+`</ansi> appears in a shower of sparks!`, user.UserId)

				user.ClearPrompt()
				return true, nil

			}
		}

		user.SendText(`No character with that name found.`)

		user.ClearPrompt()
		return true, nil

	}

	/////////////////////////
	// View characters
	/////////////////////////
	if question.Response == `view` {

		question := cmdPrompt.Ask(`Enter the name of the character you wish to view.`, []string{})
		if !question.Done {
			return true, nil
		}

		for _, char := range alts {
			if strings.EqualFold(char.Name, question.Response) {

				char.Validate()

				tmpChar := user.Character
				user.Character = &char

				Status(``, user.UserId)

				user.Character = tmpChar

				m := mobs.NewMobById(59, user.Character.RoomId)
				m.Character = char
				room.AddMob(m.InstanceId)
				m.Character.Charm(user.UserId, -1, `suicide vanish`)

				user.ClearPrompt()
				return true, nil

			}
		}

		user.SendText(`No character with that name found.`)

		user.ClearPrompt()
		return true, nil

	}

	/////////////////////////
	// Spawn a helper clone
	/////////////////////////
	if question.Response == `hire` {

		question := cmdPrompt.Ask(`Enter the name of the character you wish to hire.`, []string{})
		if !question.Done {
			return true, nil
		}

		for _, char := range alts {
			if strings.EqualFold(char.Name, question.Response) {

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
		}

		user.SendText(`No character with that name found.`)

		user.ClearPrompt()
		return true, nil

	}

	return true, nil
}
