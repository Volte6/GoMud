package spells

import "github.com/volte6/mud/util"

type SpellType int

type Spell struct {
	Name         string
	Description  string
	Type         SpellType
	MPCost       int
	WaitRounds   int
	CastFunction SpellCastFunc
}

const (
	WaitRoundsDefault = 3

	HarmfulSingle   SpellType = iota // Harmful, defaults to current aggro - magic missile etc
	HarmfulMultiple                  // Harmful, defaults to all aggro mobs - chain lightning etc
	HelpSingle                       // Helpful, defaults on self - heal etc
	HelpMultiple                     // Helpful, defaults on party - mass heal etc
)

type SpellCastFunc func(sourceUserId int, sourceMobId int, details any, cmdQueue util.CommandQueue) (util.MessageQueue, error)

var (
	SpellBook = map[string]Spell{
		"summon": {
			Name:         "Minor Heal",
			Description:  "Heals a small amount of HP.",
			Type:         HelpSingle,
			MPCost:       6,
			WaitRounds:   WaitRoundsDefault,
			CastFunction: Summon,
		},
		"mheal": {
			Name:        "Minor Heal",
			Description: "Heals a small amount of HP.",
			Type:        HelpSingle,
			MPCost:      6,
			WaitRounds:  WaitRoundsDefault,
		},
		"massheal": {
			Name:        "Minor Heal",
			Description: "Heals a small amount of HP.",
			Type:        HelpMultiple,
			MPCost:      15,
			WaitRounds:  WaitRoundsDefault,
		},
	}
)
