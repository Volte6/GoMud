package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Set(rest string, userId int) (bool, string, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("user %d not found", userId)
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {

		user.SendText(`<ansi fg="yellow-bold">description:</ansi>`)
		user.SendText(`<ansi fg="yellow">` + util.SplitStringNL(user.Character.Description, 80) + `</ansi>`)
		user.SendText(``)

		on := user.GetConfigOption(`auction`)
		onTxt := `<ansi fg="red">OFF</ansi>`
		if on == nil || on.(bool) {
			onTxt = `<ansi fg="green">ON</ansi>`
		}
		user.SendText(`<ansi fg="yellow-bold">auction:</ansi> `)
		user.SendText(onTxt)
		user.SendText(``)

		on = user.GetConfigOption(`shortadjectives`)
		onTxt = `<ansi fg="red">OFF</ansi>`
		if on == nil || on.(bool) {
			onTxt = `<ansi fg="green">ON</ansi>`
		}
		user.SendText(`<ansi fg="yellow-bold">shortadjectives:</ansi> `)
		user.SendText(onTxt)
		user.SendText(``)

		on = user.GetConfigOption(`tinymap`)
		onTxt = `<ansi fg="red">OFF</ansi>`
		if on == nil || on.(bool) {
			onTxt = `<ansi fg="green">ON</ansi>`
		}
		user.SendText(`<ansi fg="yellow-bold">tinymap:</ansi> `)
		user.SendText(onTxt)
		user.SendText(``)

		currentPrompt := user.GetConfigOption(`prompt`)
		if currentPrompt == nil {
			currentPrompt = users.PromptDefault
		}
		user.SendText(`<ansi fg="yellow-bold">prompt: </ansi> `)
		user.SendText(currentPrompt.(string))
		user.SendText(``)

		currentPrompt = user.GetConfigOption(`fprompt`)
		if currentPrompt == nil {
			currentPrompt = users.PromptDefault
		}
		user.SendText(`<ansi fg="yellow-bold">fprompt:</ansi> `)
		user.SendText(currentPrompt.(string))
		user.SendText(``)

		user.SendText(`See: <ansi fg="command">help set</ansi>`)

		return true, ``, nil
	}

	setTarget := args[0]
	args = args[1:]

	if setTarget == `description` {

		rest = strings.TrimSpace(rest[len(setTarget):])
		if len(rest) > 1024 {
			rest = rest[:1024]
		}
		user.Character.Description = rest

		user.SendText("Description set. Look at yourself to confirm.")
		return true, ``, nil
	}

	if setTarget == `auction` {
		on := user.GetConfigOption(`auction`)
		if on == nil {
			on = true
		}
		if !on.(bool) {
			on = true
			user.SendText(`Auctions turned <ansi fg="red">ON</ansi>.`)
		} else {
			on = false
			user.SendText(`Auctions turned <ansi fg="red">OFF</ansi>.`)
		}

		user.SetConfigOption(`auction`, on)

		return true, ``, nil

	}

	if setTarget == `shortadjectives` {
		on := user.GetConfigOption(`shortadjectives`)
		if on == nil || !on.(bool) {
			on = true
			user.SendText(`Short Adjectives turned <ansi fg="red">ON</ansi>.`)
		} else {
			on = false
			user.SendText(`Short Adjectives turned <ansi fg="red">OFF</ansi>.`)
		}

		user.SetConfigOption(`shortadjectives`, on)

		return true, ``, nil

	}

	if setTarget == `tinymap` {
		on := user.GetConfigOption(`tinymap`)
		if on == nil || !on.(bool) {
			on = true
			user.SendText(`Tinymap turned <ansi fg="red">ON</ansi>.`)
		} else {
			on = false
			user.SendText(`Tinymap turned <ansi fg="red">OFF</ansi>.`)
		}

		user.SetConfigOption(`tinymap`, on)

		return true, ``, nil

	}

	if setTarget == `prompt` {

		if len(args) < 1 {
			currentPrompt := user.GetConfigOption(`prompt`)
			if currentPrompt == nil {
				currentPrompt = users.PromptDefault
			}
			user.SendText("Your current prompt:\n")
			user.SendText(currentPrompt.(string))
			user.SendText("\n" + `Type <ansi fg="command">help set-prompt</ansi> for more info on customizing prompts.` + "\n")
			return true, ``, nil
		}

		promptStr := rest[len(setTarget)+1:]

		if promptStr == `default` {
			user.SetConfigOption(`prompt`, nil)
			user.SetConfigOption(`prompt-compiled`, nil)
			user.SetConfigOption(`prompt`, users.PromptDefault)
			user.SetConfigOption(`prompt-compiled`, util.ConvertColorShortTags(users.PromptDefault))
		} else if promptStr == `none` {
			user.SetConfigOption(`prompt`, ``)
			user.SetConfigOption(`prompt-compiled`, ``)
		} else {
			user.SetConfigOption(`prompt`, promptStr)
			user.SetConfigOption(`prompt-compiled`, util.ConvertColorShortTags(promptStr))
		}

		user.SendText("Prompt set.")
		return true, ``, nil

	}

	if setTarget == `fprompt` {

		if len(args) < 1 {
			currentPrompt := user.GetConfigOption(`fprompt`)
			if currentPrompt == nil {
				currentPrompt = users.PromptDefault
			}
			user.SendText("Your current fprompt:\n")
			user.SendText(currentPrompt.(string))
			user.SendText("\n" + `Type <ansi fg="command">help set-prompt</ansi> for more info on customizing prompts.` + "\n")
			return true, ``, nil
		}

		promptStr := rest[len(setTarget)+1:]

		if promptStr == `default` {
			user.SetConfigOption(`fprompt`, nil)
			user.SetConfigOption(`fprompt-compiled`, nil)
		} else if promptStr == `none` {
			user.SetConfigOption(`fprompt`, ``)
			user.SetConfigOption(`fprompt-compiled`, ``)
		} else {
			user.SetConfigOption(`fprompt`, promptStr)
			user.SetConfigOption(`fprompt-compiled`, util.ConvertColorShortTags(promptStr))
		}

		user.SendText("fprompt set.")
		return true, ``, nil

	}

	// Are they setting a macro?
	if len(setTarget) == 2 && setTarget[0] == '=' {
		macroNum, _ := strconv.Atoi(string(args[0][1]))
		if macroNum == 0 {
			user.SendText("Invalid macro number supplied.")
			return true, ``, nil
		}
		if user.Macros == nil {
			user.Macros = make(map[string]string)
		}
		rest = strings.TrimSpace(rest[2:])

		if len(rest) > 128 {
			rest = rest[:128]
		}

		if len(rest) == 0 {
			delete(user.Macros, args[0])

			user.SendText(
				fmt.Sprintf(`Macro <ansi fg="command">=%d</ansi> deleted.`, macroNum),
			)
		} else {

			for _, cmd := range strings.Split(rest, ";") {
				if len(cmd) > 0 {
					if cmd[0] == '=' {
						user.SendText(
							`You cannot reference macros inside of a macro`,
						)
						return true, ``, nil
					}
				}
			}

			user.Macros[args[0]] = rest

			user.SendText(
				fmt.Sprintf(`Macro set. Type <ansi fg="command">=%d</ansi> or press <ansi fg="command">F%d</ansi> to use it.`, macroNum, macroNum),
			)
		}

		return true, ``, nil
	}
	// Setting macros:
	// set =1 "say hello"

	return true, ``, nil
}
