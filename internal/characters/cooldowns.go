package characters

import (
	"github.com/GoMudEngine/GoMud/internal/gametime"
)

type Cooldowns map[string]int

func (cd Cooldowns) RoundTick() {
	for trackingTag, _ := range cd {
		cd[trackingTag] = cd[trackingTag] - 1
	}
}

func (cd Cooldowns) Prune() {
	for trackingTag, cooldownRounds := range cd {
		if cooldownRounds <= 0 {
			delete(cd, trackingTag)
		}
	}
}

func (cd Cooldowns) Try(trackingTag string, cooldownPeriod string) bool {
	if cd == nil {
		cd = make(Cooldowns)
	}

	cd.Prune()

	cooldownRounds := int(gametime.GetDate(1000000).AddPeriod(cooldownPeriod) - 1000000)

	if cooldownRounds < 1 {
		return true
	}

	if _, ok := cd[trackingTag]; ok {
		if cd[trackingTag] > 0 {
			return false
		}
	}

	cd[trackingTag] = cooldownRounds
	return true
}
