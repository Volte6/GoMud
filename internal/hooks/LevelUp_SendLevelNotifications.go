package hooks

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

func SendLevelNotifications(e events.Event) bool {

	evt, typeOk := e.(events.LevelUp)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "LevelUp", "Actual Type", e.Type())
		return false
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return true
	}

	levelUpData := map[string]interface{}{
		"level":          evt.NewLevel,
		"statsDelta":     evt.StatsDelta,
		"trainingPoints": evt.TrainingPoints,
		"statPoints":     evt.StatPoints,
		"livesUp":        evt.LivesGained,
	}
	levelUpStr, _ := templates.Process("character/levelup", levelUpData)

	user.SendText(levelUpStr)

	user.PlaySound(`levelup`, `other`)

	events.AddToQueue(events.Broadcast{
		Text: fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> <ansi fg="yellow">has reached level %d!</ansi> <ansi fg="magenta-bold">***</ansi>%s`, evt.CharacterName, evt.NewLevel, term.CRLFStr),
	})

	return true
}
