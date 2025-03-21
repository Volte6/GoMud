package hooks

import (
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

// Checks whether their level is too high for a guide
func RedrawPrompt_SendRedraw(e events.Event) events.EventReturn {

	evt, typeOk := e.(events.RedrawPrompt)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "RedrawPrompt", "Actual Type", e.Type())
		return events.Cancel
	}

	if user := users.GetByUserId(evt.UserId); user != nil {

		newCmdPrompt := user.GetCommandPrompt()

		if evt.OnlyIfChanged {

			oldCmdPrompt := user.GetTempData(`cmdprompt`)

			// If the prompt hasn't changed, skip redrawing
			if oldCmdPrompt != nil && oldCmdPrompt.(string) == newCmdPrompt {
				return events.Continue
			}

			// save the new prompt for next time we want to check
			user.SetTempData(`cmdprompt`, newCmdPrompt)

		}

		pTxt := templates.AnsiParse(newCmdPrompt)
		connections.SendTo([]byte(pTxt), user.ConnectionId())

	}

	return events.Continue
}
