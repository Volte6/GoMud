package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/templates"
)

//
// Watches the rounds go by
// Performs auction status updates
//

func SunriseSunset(e events.Event) bool {

	evt := e.(events.NewRound)

	gdBefore := gametime.GetDate(evt.RoundNumber - 1)

	gdNow := gametime.GetDate()

	if gdBefore.Night != gdNow.Night {
		if gdNow.Night {
			sunsetTxt, _ := templates.Process("generic/sunset", nil)

			events.AddToQueue(events.Broadcast{
				Text: sunsetTxt,
			})

		} else {
			sunriseTxt, _ := templates.Process("generic/sunrise", gdNow)
			events.AddToQueue(events.Broadcast{
				Text: sunriseTxt,
			})

		}
	}

	return true
}
