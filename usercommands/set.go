package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Set(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {

		response.SendUserMessage(userId, `<ansi fg="yellow-bold">description:</ansi>`, true)
		response.SendUserMessage(userId, `<ansi fg="yellow">`+util.SplitStringNL(user.Character.Description, 80)+`</ansi>`, true)
		response.SendUserMessage(userId, ``, true)

		on := user.GetConfigOption(`auction`)
		onTxt := `<ansi fg="red">OFF</ansi>`
		if on == nil || on.(bool) {
			onTxt = `<ansi fg="green">ON</ansi>`
		}
		response.SendUserMessage(userId, `<ansi fg="yellow-bold">auction:</ansi> `, true)
		response.SendUserMessage(userId, onTxt, true)
		response.SendUserMessage(userId, ``, true)

		on = user.GetConfigOption(`shortadjectives`)
		onTxt = `<ansi fg="red">OFF</ansi>`
		if on == nil || on.(bool) {
			onTxt = `<ansi fg="green">ON</ansi>`
		}
		response.SendUserMessage(userId, `<ansi fg="yellow-bold">shortadjectives:</ansi> `, true)
		response.SendUserMessage(userId, onTxt, true)
		response.SendUserMessage(userId, ``, true)

		on = user.GetConfigOption(`tinymap`)
		onTxt = `<ansi fg="red">OFF</ansi>`
		if on == nil || on.(bool) {
			onTxt = `<ansi fg="green">ON</ansi>`
		}
		response.SendUserMessage(userId, `<ansi fg="yellow-bold">tinymap:</ansi> `, true)
		response.SendUserMessage(userId, onTxt, true)
		response.SendUserMessage(userId, ``, true)

		currentPrompt := user.GetConfigOption(`prompt`)
		if currentPrompt == nil {
			currentPrompt = users.PromptDefault
		}
		response.SendUserMessage(userId, `<ansi fg="yellow-bold">prompt: </ansi> `, true)
		response.SendUserMessage(userId, currentPrompt.(string), true)
		response.SendUserMessage(userId, ``, true)

		currentPrompt = user.GetConfigOption(`fprompt`)
		if currentPrompt == nil {
			currentPrompt = users.PromptDefault
		}
		response.SendUserMessage(userId, `<ansi fg="yellow-bold">fprompt:</ansi> `, true)
		response.SendUserMessage(userId, currentPrompt.(string), true)
		response.SendUserMessage(userId, ``, true)

		response.SendUserMessage(userId, `See: <ansi fg="command">help set</ansi>`, true)

		response.Handled = true
		return response, nil
	}

	setTarget := args[0]
	args = args[1:]

	if setTarget == `description` {

		rest = strings.TrimSpace(rest[len(setTarget):])
		if len(rest) > 1024 {
			rest = rest[:1024]
		}
		user.Character.Description = rest

		response.SendUserMessage(userId, "Description set. Look at yourself to confirm.", true)
		response.Handled = true
		return response, nil
	}

	if setTarget == `auction` {
		on := user.GetConfigOption(`auction`)
		if on == nil {
			on = true
		}
		if !on.(bool) {
			on = true
			response.SendUserMessage(userId, `Auctions turned <ansi fg="red">ON</ansi>.`, true)
		} else {
			on = false
			response.SendUserMessage(userId, `Auctions turned <ansi fg="red">OFF</ansi>.`, true)
		}

		user.SetConfigOption(`auction`, on)

		response.Handled = true
		return response, nil

	}

	if setTarget == `shortadjectives` {
		on := user.GetConfigOption(`shortadjectives`)
		if on == nil || !on.(bool) {
			on = true
			response.SendUserMessage(userId, `Short Adjectives turned <ansi fg="red">ON</ansi>.`, true)
		} else {
			on = false
			response.SendUserMessage(userId, `Short Adjectives turned <ansi fg="red">OFF</ansi>.`, true)
		}

		user.SetConfigOption(`shortadjectives`, on)

		response.Handled = true
		return response, nil

	}

	if setTarget == `tinymap` {
		on := user.GetConfigOption(`tinymap`)
		if on == nil || !on.(bool) {
			on = true
			response.SendUserMessage(userId, `Tinymap turned <ansi fg="red">ON</ansi>.`, true)
		} else {
			on = false
			response.SendUserMessage(userId, `Tinymap turned <ansi fg="red">OFF</ansi>.`, true)
		}

		user.SetConfigOption(`tinymap`, on)

		response.Handled = true
		return response, nil

	}

	if setTarget == `prompt` {

		if len(args) < 1 {
			currentPrompt := user.GetConfigOption(`prompt`)
			if currentPrompt == nil {
				currentPrompt = users.PromptDefault
			}
			response.SendUserMessage(userId, "Your current prompt:\n", true)
			response.SendUserMessage(userId, currentPrompt.(string), true)
			response.SendUserMessage(userId, "\n"+`Type <ansi fg="command">help set-prompt</ansi> for more info on customizing prompts.`+"\n", true)
			response.Handled = true
			return response, nil
		}

		promptStr := rest[len(setTarget)+1:]

		if promptStr == `default` {
			user.SetConfigOption(`prompt`, nil)
			user.SetConfigOption(`prompt-compiled`, nil)
			user.SetConfigOption(`prompt`, users.PromptDefault)
			user.SetConfigOption(`prompt-compiled`, users.CompilePrompt(users.PromptDefault))
		} else if promptStr == `none` {
			user.SetConfigOption(`prompt`, ``)
			user.SetConfigOption(`prompt-compiled`, ``)
		} else {
			user.SetConfigOption(`prompt`, promptStr)
			user.SetConfigOption(`prompt-compiled`, users.CompilePrompt(promptStr))
		}

		response.SendUserMessage(userId, "Prompt set.", true)
		response.Handled = true
		return response, nil

	}

	if setTarget == `fprompt` {

		if len(args) < 1 {
			currentPrompt := user.GetConfigOption(`fprompt`)
			if currentPrompt == nil {
				currentPrompt = users.PromptDefault
			}
			response.SendUserMessage(userId, "Your current fprompt:\n", true)
			response.SendUserMessage(userId, currentPrompt.(string), true)
			response.SendUserMessage(userId, "\n"+`Type <ansi fg="command">help set-prompt</ansi> for more info on customizing prompts.`+"\n", true)
			response.Handled = true
			return response, nil
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
			user.SetConfigOption(`fprompt-compiled`, users.CompilePrompt(promptStr))
		}

		response.SendUserMessage(userId, "fprompt set.", true)
		response.Handled = true
		return response, nil

	}

	// Are they setting a macro?
	if len(setTarget) == 2 && setTarget[0] == '=' {
		macroNum, _ := strconv.Atoi(string(args[0][1]))
		if macroNum == 0 {
			response.SendUserMessage(userId, "Invalid macro number supplied.", true)
			response.Handled = true
			return response, nil
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

			response.SendUserMessage(userId,
				fmt.Sprintf(`Macro <ansi fg="command">=%d</ansi> deleted.`, macroNum),
				true)
		} else {

			for _, cmd := range strings.Split(rest, ";") {
				if len(cmd) > 0 {
					if cmd[0] == '=' {
						response.SendUserMessage(userId,
							`You cannot reference macros inside of a macro`,
							true)
						response.Handled = true
						return response, nil
					}
				}
			}

			user.Macros[args[0]] = rest

			response.SendUserMessage(userId,
				fmt.Sprintf(`Macro set. Type <ansi fg="command">=%d</ansi> or press <ansi fg="command">F%d</ansi> to use it.`, macroNum, macroNum),
				true)
		}

		response.Handled = true
		return response, nil
	}
	// Setting macros:
	// set =1 "say hello"

	response.Handled = true
	return response, nil
}
