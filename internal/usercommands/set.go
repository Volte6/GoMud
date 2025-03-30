package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Set(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))
	c := configs.GetTextFormatsConfig()

	if len(args) == 0 {

		user.SendText(`<ansi fg="yellow-bold">description:</ansi>`)
		user.SendText(`<ansi fg="yellow">` + util.SplitStringNL(user.Character.Description, 80) + `</ansi>`)
		user.SendText(``)

		user.SendText(`<ansi fg="yellow-bold">ScreenReader:</ansi> `)
		if user.ScreenReader {
			user.SendText(`<ansi fg="green">ON</ansi>`)
		} else {
			user.SendText(`<ansi fg="red">OFF</ansi>`)
		}
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
			currentPrompt = c.Prompt.String()
		}
		user.SendText(`<ansi fg="yellow-bold">prompt: </ansi> `)
		user.SendText(currentPrompt.(string))
		user.SendText(``)

		currentPrompt = user.GetConfigOption(`fprompt`)
		if currentPrompt == nil {
			currentPrompt = c.Prompt.String()
		}
		user.SendText(`<ansi fg="yellow-bold">fprompt:</ansi> `)
		user.SendText(currentPrompt.(string))
		user.SendText(``)

		currentWimpy := user.GetConfigOption(`wimpy`)
		if currentWimpy == nil {
			currentWimpy = 0
		}
		user.SendText(`<ansi fg="yellow-bold">wimpy:</ansi> `)
		user.SendText(fmt.Sprintf(`%d%%`, currentWimpy.(int)))
		user.SendText(``)

		user.SendText(`See: <ansi fg="command">help set</ansi>`)

		return true, nil
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

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `description`,
		})

		return true, nil
	}

	if setTarget == `auction` {
		on := user.GetConfigOption(`auction`)
		if on == nil {
			on = true
		}
		if !on.(bool) {
			on = true
			user.SendText(`Auctions toggled <ansi fg="red">ON</ansi>.`)
		} else {
			on = false
			user.SendText(`Auctions toggled <ansi fg="red">OFF</ansi>.`)
		}

		user.SetConfigOption(`auction`, on)

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `auction`,
		})

		return true, nil

	}

	if setTarget == `shortadjectives` {
		on := user.GetConfigOption(`shortadjectives`)
		if on == nil {
			on = false
		}
		if !on.(bool) {
			on = true
			user.SendText(`Short Adjectives toggled <ansi fg="red">ON</ansi>.`)
		} else {
			on = false
			user.SendText(`Short Adjectives toggled <ansi fg="red">OFF</ansi>.`)
		}

		user.SetConfigOption(`shortadjectives`, on)

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `shortadjectives`,
		})

		return true, nil

	}

	if setTarget == `tinymap` {
		on := user.GetConfigOption(`tinymap`)
		if on == nil {
			on = true
		}
		if !on.(bool) {
			on = true
			user.SendText(`Tinymap toggled <ansi fg="red">ON</ansi>.`)
		} else {
			on = false
			user.SendText(`Tinymap toggled <ansi fg="red">OFF</ansi>.`)
		}

		user.SetConfigOption(`tinymap`, on)

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `tinymap`,
		})

		return true, nil

	}

	if setTarget == `prompt` {

		if len(args) < 1 {
			currentPrompt := user.GetConfigOption(`prompt`)
			if currentPrompt == nil {
				currentPrompt = c.Prompt.String()
			}
			user.SendText("Your current prompt:\n")
			user.SendText(currentPrompt.(string))
			user.SendText("\n" + `Type <ansi fg="command">help set-prompt</ansi> for more info on customizing prompts.` + "\n")
			return true, nil
		}

		promptStr := rest[len(setTarget)+1:]

		if promptStr == `default` {
			user.SetConfigOption(`prompt`, nil)
			user.SetConfigOption(`prompt-compiled`, nil)
			user.SetConfigOption(`prompt`, c.Prompt.String())
			user.SetConfigOption(`prompt-compiled`, util.ConvertColorShortTags(c.Prompt.String()))
		} else if promptStr == `none` {
			user.SetConfigOption(`prompt`, ``)
			user.SetConfigOption(`prompt-compiled`, ``)
		} else {
			user.SetConfigOption(`prompt`, promptStr)
			user.SetConfigOption(`prompt-compiled`, util.ConvertColorShortTags(promptStr))
		}

		user.SendText("Prompt set.")

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `prompt`,
		})

		return true, nil

	}

	if setTarget == `fprompt` {

		if len(args) < 1 {
			currentPrompt := user.GetConfigOption(`fprompt`)
			if currentPrompt == nil {
				currentPrompt = c.Prompt.String()
			}
			user.SendText("Your current fprompt:\n")
			user.SendText(currentPrompt.(string))
			user.SendText("\n" + `Type <ansi fg="command">help set-prompt</ansi> for more info on customizing prompts.` + "\n")
			return true, nil
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

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `fprompt`,
		})

		return true, nil

	}

	if setTarget == `wimpy` {

		if len(args) < 1 {
			currentWimpy := user.GetConfigOption(`wimpy`)
			if currentWimpy == nil {
				currentWimpy = 0
			}
			user.SendText("Your current wimpy:\n")
			user.SendText(fmt.Sprintf(`%d%%`, currentWimpy.(int)))
			user.SendText("\n" + `Type <ansi fg="command">help wimpy</ansi> to learn about the wimpy setting.` + "\n")
			return true, nil
		}

		wimpyStr := rest[len(setTarget)+1:]
		wimipyInt, _ := strconv.Atoi(wimpyStr)

		if wimipyInt == 0 {
			user.SetConfigOption(`wimpy`, nil)
		} else {
			user.SetConfigOption(`wimpy`, wimipyInt)
		}

		user.SendText("wimpy set.")

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `wimpy`,
		})

		return true, nil

	}

	if setTarget == `screenreader` {
		if user.ScreenReader {
			user.SendText(`ScreenReader mode toggled <ansi fg="red">OFF</ansi>.`)
		} else {
			user.SendText(`ScreenReader mode toggled <ansi fg="red">ON</ansi>.`)
		}
		user.ScreenReader = !user.ScreenReader

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `screenreader`,
		})

		return true, nil
	}

	// hidden command for debugging purposes.
	if setTarget == `gmcp` {

		cs := connections.GetClientSettings(user.ConnectionId())
		if cs.GMCPModules == nil {
			cs.GMCPModules = map[string]int{}
		}

		if _, ok := cs.GMCPModules[`*`]; ok {
			user.SendText(`GMCP forced support toggled <ansi fg="red">OFF</ansi>.`)
			delete(cs.GMCPModules, `*`)
		} else {
			user.SendText(`GMCP forced support toggled <ansi fg="red">ON</ansi>.`)
			cs.GMCPModules[`*`] = 1
		}
		connections.OverwriteClientSettings(user.ConnectionId(), cs)

		return true, nil
	}

	// Are they setting a macro? // setTarget should be "=1" etc
	if len(setTarget) == 2 && setTarget[0] == '=' {

		setVal := strings.Join(args, ` `)

		macroNum, _ := strconv.Atoi(string(setTarget[1]))
		if macroNum == 0 {
			user.SendText("Invalid macro number supplied.")
			return true, nil
		}

		if user.Macros == nil {
			user.Macros = make(map[string]string)
		}

		// Keep macros small enough.
		if len(setVal) > 128 {
			setVal = setVal[:128]
		}

		if len(setVal) == 0 {
			delete(user.Macros, setTarget)
			user.SendText(fmt.Sprintf(`Macro <ansi fg="command">=%d</ansi> deleted.`, macroNum))
		} else {

			allComands := strings.Split(setVal, ";")
			if len(allComands) > 10 {
				user.SendText(`Macros are limited to 10 commands.`)
				return true, nil
			}

			finalMacroCommands := []string{}

			for _, cmd := range allComands {

				if len(cmd) > 0 {
					if cmd[0] == '=' {
						user.SendText(`You cannot reference macros inside of a macro`)
						return true, nil
					}
					finalMacroCommands = append(finalMacroCommands, cmd)
				}

			}

			if len(finalMacroCommands) < 1 {
				user.SendText(`There was a problem setting your macro.`)
				return true, nil
			}

			user.Macros[setTarget] = strings.Join(finalMacroCommands, `;`)

			user.SendText(fmt.Sprintf(`Macro set. Type <ansi fg="command">=%d</ansi> or (if your terminal supports it) press <ansi fg="command">F%d</ansi> to use it.`, macroNum, macroNum))
		}

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `macro`,
		})

		return true, nil
	}

	return true, nil
}
