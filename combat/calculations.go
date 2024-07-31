package combat

import (
	"github.com/volte6/mud/characters"
)

func PowerRanking(atkChar characters.Character, defChar characters.Character) float64 {

	attacks, dCount, dSides, dBonus, _ := atkChar.Equipment.Weapon.GetDiceRoll()
	atkDmg := attacks * (dCount*dSides + dBonus)

	attacks, dCount, dSides, dBonus, _ = defChar.Equipment.Weapon.GetDiceRoll()
	defDmg := attacks * (dCount*dSides + dBonus)

	pct := 0.0
	if defDmg == 0 {
		pct += 0.4
	} else {
		pct += 0.4 * float64(atkDmg) / float64(defDmg)
	}

	if defChar.Stats.Speed.ValueAdj == 0 {
		pct += 0.3
	} else {
		pct += 0.3 * float64(atkChar.Stats.Speed.ValueAdj) / float64(defChar.Stats.Speed.ValueAdj)
	}

	if defChar.HealthMax.Value == 0 {
		pct += 0.2
	} else {
		pct += 0.2 * float64(atkChar.HealthMax.Value) / float64(defChar.HealthMax.Value)
	}

	if defChar.GetDefense() == 0 {
		pct += 0.1
	} else {
		pct += 0.1 * float64(atkChar.GetDefense()) / float64(defChar.GetDefense())
	}

	return pct
}
