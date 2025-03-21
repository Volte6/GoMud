package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Status(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	//possibleStatuses := []string{`strength`, `speed`, `smarts`, `vitality`, `mysticism`, `perception`}

	if rest != `` {

		if rest != `train` {
			user.SendText("status WHAT???")
			return true, nil
		}

		user.DidTip(`status train`, true)

		cmdPrompt, isNew := user.StartPrompt(`status`, rest)

		if isNew {
			tplTxt, _ := templates.Process("character/status-train", user)
			user.SendText(tplTxt)
		}

		question := cmdPrompt.Ask(`Increase which?`, []string{`strength`, `speed`, `smarts`, `vitality`, `mysticism`, `perception`, `quit`}, `quit`)
		if !question.Done {
			return true, nil
		}

		if question.Response == `quit` {
			user.ClearPrompt()
			return true, nil
		}

		match, closeMatch := util.FindMatchIn(question.Response, []string{`strength`, `speed`, `smarts`, `vitality`, `mysticism`, `perception`}...)

		question.RejectResponse() // Always reset this question, since we want to keep reusing it.

		if user.Character.StatPoints < 1 {
			user.SendText(`Oops! You have no stat points to spend!`)
			user.ClearPrompt()
			return true, nil
		}
		selection := match
		if match == `` {
			selection = closeMatch
		}

		before := 0
		after := 0
		spent := 0

		switch selection {
		case `strength`:
			before = user.Character.Stats.Strength.Training
			user.Character.Stats.Strength.Training += 1
			spent = 1
		case `speed`:
			before = user.Character.Stats.Speed.Training
			user.Character.Stats.Speed.Training += 1
			spent = 1
		case `smarts`:
			before = user.Character.Stats.Smarts.Training
			user.Character.Stats.Smarts.Training += 1
			spent = 1
		case `vitality`:
			before = user.Character.Stats.Vitality.Training
			user.Character.Stats.Vitality.Training += 1
			spent = 1
		case `mysticism`:
			before = user.Character.Stats.Mysticism.Training
			user.Character.Stats.Mysticism.Training += 1
			spent = 1
		case `perception`:
			before = user.Character.Stats.Perception.Training
			user.Character.Stats.Perception.Training += 1
			spent = 1
		}

		if spent > 0 {
			after = before + 1
			user.Character.StatPoints -= 1

			user.Character.Validate()

			user.SendText(
				fmt.Sprintf(term.CRLFStr+`<ansi fg="210">Your <ansi fg="yellow">%s</ansi> training improves from <ansi fg="201">%d</ansi> to <ansi fg="201">%d</ansi>!</ansi>`, selection, before, after))
		}

		tplTxt, _ := templates.Process("character/status-train", user)

		if spent > 0 {
			tplTxt = strings.Replace(tplTxt, `fakeprop="`+selection+`"`, `bg="highlight"`, 1)
		}

		user.SendText(tplTxt)

		return true, nil
	}

	tplTxt, _ := templates.Process("character/status", user)
	user.SendText(tplTxt)

	Inventory(``, user, room, flags)

	return true, nil
}
