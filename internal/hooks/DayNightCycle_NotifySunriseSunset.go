package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/gametime"
	"github.com/GoMudEngine/GoMud/internal/templates"
)

//
// Watches the rounds go by
// Spawns the loot goblin every so often
//

func NotifySunriseSunset(e events.Event) events.ListenerReturn {
	evt, typeOk := e.(events.DayNightCycle)
	if !typeOk {
		return events.Cancel
	}

	if evt.IsSunrise {

		sunriseTxt, _ := templates.Process("generic/sunrise", gametime.GetDate())
		sunriseTxtSR, _ := templates.Process("generic/sunrise", gametime.GetDate(), templates.ForceScreenReaderUserId)

		events.AddToQueue(events.Broadcast{
			Text:             sunriseTxt,
			TextScreenReader: sunriseTxtSR,
		})
		return events.Continue
	}

	sunsetTxt, _ := templates.Process("generic/sunset", nil)
	sunsetTxtSR, _ := templates.Process("generic/sunset", nil, templates.ForceScreenReaderUserId)

	events.AddToQueue(events.Broadcast{
		Text:             sunsetTxt,
		TextScreenReader: sunsetTxtSR,
	})

	return events.Continue
}
