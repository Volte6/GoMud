package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/templates"
)

//
// Watches the rounds go by
// Spawns the loot goblin every so often
//

func NotifySunriseSunset(e events.Event) bool {
	evt, typeOk := e.(events.DayNightCycle)
	if !typeOk {
		return false
	}

	if evt.IsSunrise {
		sunriseTxt, _ := templates.Process("generic/sunrise", gametime.GetDate())
		events.AddToQueue(events.Broadcast{
			Text: sunriseTxt,
		})
		return true
	}

	sunsetTxt, _ := templates.Process("generic/sunset", nil)
	events.AddToQueue(events.Broadcast{
		Text: sunsetTxt,
	})

	return true
}
