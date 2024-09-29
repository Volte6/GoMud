package combat

import (
	"math"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/users"
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

func ChanceToTame(s *users.UserRecord, t *mobs.Mob) int {

	var MOD_SKILL_MIN int = 1   // Minimum base tame ability
	var MOD_SKILL_MAX int = 100 // Maximum base tame ability

	var MOD_SIZE_SMALL int = 0    // Modifier for small creatures
	var MOD_SIZE_MEDIUM int = -10 // Modifier for medium creatures
	var MOD_SIZE_LARGE int = -25  // Modifier for large creatures

	var MOD_LEVELDIFF_MIN int = -25 // Lowest level delta modifier
	var MOD_LEVELDIFF_MAX int = 25  // Highest level delta modifier

	var MOD_HEALTHPERCENT_MAX float64 = 50 // Highest possible bonus for target HP being reduced

	var FACTOR_IS_AGGRO float64 = .50 // Overall reduction of chance if target is aggro

	proficiencyModifier := s.Character.MobMastery.GetTame(int(t.MobId))

	if proficiencyModifier < MOD_SKILL_MIN {
		proficiencyModifier = MOD_SKILL_MIN
	} else if proficiencyModifier > MOD_SKILL_MAX {
		proficiencyModifier = MOD_SKILL_MAX
	}

	raceInfo := races.GetRace(s.Character.RaceId)

	sizeModifier := 0
	switch raceInfo.Size {
	case races.Large:
		sizeModifier = MOD_SIZE_LARGE
	case races.Small:
		sizeModifier = MOD_SIZE_SMALL
	case races.Medium:
	default:
		sizeModifier = MOD_SIZE_MEDIUM
	}

	levelDiff := s.Character.Level - t.Character.Level
	if levelDiff > MOD_LEVELDIFF_MAX {
		levelDiff = MOD_LEVELDIFF_MAX
	} else if levelDiff < MOD_LEVELDIFF_MIN {
		levelDiff = MOD_LEVELDIFF_MIN
	}

	healthModifier := MOD_HEALTHPERCENT_MAX - math.Ceil(float64(s.Character.Health)/float64(s.Character.HealthMax.Value)*MOD_HEALTHPERCENT_MAX)

	var aggroModifier float64 = 1
	if t.Character.IsAggro(s.UserId, 0) {
		aggroModifier = FACTOR_IS_AGGRO
	}

	return int(math.Ceil((float64(proficiencyModifier) + float64(levelDiff) + healthModifier + float64(sizeModifier)) * aggroModifier))
}

// Evil killing good:
// -50 - 50 = -100
// -100 - 50 = -150

// Good killing evil:
// 50 - -50 = 100
// 100 - -50 = 150
func AlignmentChange(killerAlignment int8, killedAlignment int8) int {

	delta := killerAlignment - killedAlignment
	deltaAbs := math.Abs(float64(delta))

	factor := 0
	if delta < 0 { // if killer alignment was less than killed alignment
		factor = 1
	} else if delta > 0 { // opposite
		factor = -1
	}

	if deltaAbs < 10 {
		return 0
	}

	if deltaAbs < 30 {
		return factor * 1
	}

	if deltaAbs < 60 {
		return factor * 2
	}

	if deltaAbs < 80 {
		return factor * 3
	}

	return factor * 4
}
