package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/gametime"
)

//
// Watches the rounds go by
// fires event when it's a new day
//

func CheckNewDay(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	gdBefore := gametime.GetDate(evt.RoundNumber - 1)

	gdNow := gametime.GetDate()

	if gdBefore.Night != gdNow.Night {

		events.AddToQueue(events.DayNightCycle{
			IsSunrise: !gdNow.Night,
			Day:       gdNow.Day,
			Month:     gdNow.Month,
			Year:      gdNow.Year,
			Time:      gdNow.String(),
		})

	}

	return events.Continue
}
