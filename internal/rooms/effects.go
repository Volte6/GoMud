package rooms

import "github.com/volte6/gomud/internal/util"

type AreaEffect struct {
	Type         EffectType `yaml:"-"`
	StartedRound uint64     `yaml:"-"`
}

func (a *AreaEffect) Expired() bool {
	return a.RoundsLeft() == 0
}

func (a *AreaEffect) Cooling() bool {
	return a.RoundsLeft() < 0
}

func (a *AreaEffect) RoundsLeft() uint64 {

	details, ok := effectDetails[a.Type]
	if !ok {
		return 0
	}

	return a.StartedRound + uint64(details.LifeRounds) - util.GetRoundCount()
}

func (a *AreaEffect) Prunable() bool {

	details, ok := effectDetails[a.Type]
	if !ok {
		return true
	}

	return util.GetRoundCount()-a.StartedRound >= uint64(details.CooldownRounds)

}

func NewEffect(eType EffectType) AreaEffect {
	return AreaEffect{
		Type:         eType,
		StartedRound: util.GetRoundCount(),
	}
}

type EffectType string

type EffectDetails struct {
	Type           EffectType // The identifier
	Description    string     // A brief description
	LifeRounds     int        // How many rounds does it last
	CooldownRounds int        // how long before it can be allowed to be re-applied
}

var (
	Wildfire EffectType = `wildfire`

	effectDetails = map[EffectType]EffectDetails{
		Wildfire: EffectDetails{
			Type:           Wildfire,
			Description:    `A spreading fire, burning through the area.`,
			LifeRounds:     6,
			CooldownRounds: 10,
		},
	}
)
