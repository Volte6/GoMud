package skills

import (
	"strings"
)

/****************
	dual-wield
	enchant
	inspect
	map
	peep
	portal
	scribe
	search
	skulduggery
	track
	brawling
	protection
	herbalism
	  - forage (find ingredients, special ones in each zone?)
	  - mix (potions?)
	alchemy
	  - alchemize (turn items to gold)
	  - combine (combine two items into one new item, retaining random stats from each)
	beastmastery
	  - call (calls a random beast/pet to follow you around) - higher level gets more interesting stuff
	  - command (hunt) - find players, mobs, or items nearby
	  - command (attack/break) - attack a player or mob
	  - command (guard) - absorbs damage the player would have taken
	healing
	  - heal (minor heal self or others)
	  - cure (cure poison)
	  - majorheal (major heal self or others)
	  -





 ****************/

type SkillTag string

func (s SkillTag) String(subtag ...string) string {
	result := string(s)
	if len(subtag) > 0 {
		result += `:` + strings.Join(subtag, `:`)
	}
	return result
}

func (s SkillTag) Sub(subtag string) SkillTag {
	return SkillTag(string(s) + subtag)
}

const (
	DualWield   SkillTag = `dual-wield`
	Map         SkillTag = `map`
	Enchant     SkillTag = `enchant`
	Peep        SkillTag = `peep`
	Inspect     SkillTag = `inspect`
	Portal      SkillTag = `portal`
	Script      SkillTag = `scribe`
	Search      SkillTag = `search`
	Track       SkillTag = `track`
	Skulduggery SkillTag = `skulduggery`
	Brawling    SkillTag = `brawling`
	Scribe      SkillTag = `scribe`
	Protection  SkillTag = `protection`
)

var (
	Professions = map[string][]SkillTag{
		"treasure hunter": {
			Map,
			Search,
			Peep,
			Inspect,
		},
		"assassin": {
			DualWield,
			Skulduggery,
			Track,
		},
		"explorer": {
			Portal,
			Map,
			Scribe,
		},
		"arcane scholar": {
			Enchant,
			Scribe,
			Inspect,
		},
		"warrior": {
			DualWield,
			Brawling,
		},
		"paladin": {
			Protection,
			Brawling,
		},
	}
)

type ProfessionRank struct {
	Profession       string
	ExperienceTitle  string
	TotalPointsSpent float64
	PointsToMax      float64
	Completion       float64
	Skills           []string
}

func GetProfessionRanks(allRanks map[string]int) []ProfessionRank {

	professionList := []ProfessionRank{}

	for professionName, skills := range Professions {

		ranking := ProfessionRank{Profession: professionName}

		for _, skillName := range skills {

			skillLevel := 0
			if rankVal, ok := allRanks[string(skillName)]; ok {
				skillLevel = rankVal
			}
			if skillLevel > 4 {
				skillLevel = 4
			}
			totalSkill := (skillLevel * (skillLevel + 1)) / 2

			ranking.PointsToMax += 10.0 // Each skill has 4 levels, so possible 10 points per skill
			ranking.TotalPointsSpent += float64(totalSkill)
			ranking.Skills = append(ranking.Skills, string(skillName))
		}

		ranking.Completion = ranking.TotalPointsSpent / ranking.PointsToMax
		ranking.ExperienceTitle = GetExperienceLevel(ranking.Completion)

		professionList = append(professionList, ranking)
	}

	return professionList
}

func GetProfession(allRanks map[string]int) string {

	rankData := GetProfessionRanks(allRanks)

	var highestCompletion float64 = 0
	//var highestSpend float64 = 0
	chosenProfessions := []string{}
	experienceName := ``

	for _, pRank := range rankData {

		if pRank.Completion == 0 {
			continue
		}

		if pRank.Completion > highestCompletion {
			highestCompletion = pRank.Completion
			//highestSpend = pRank.TotalPointsSpent
			chosenProfessions = []string{}
		}

		if pRank.Completion == highestCompletion {
			experienceName = pRank.ExperienceTitle
			chosenProfessions = append(chosenProfessions, pRank.Profession)
		}
	}

	if len(chosenProfessions) < 1 {
		return `scrub`
	}

	if len(experienceName) > 0 {
		experienceName = experienceName + ` `
	}

	return experienceName + strings.Join(chosenProfessions, `/`)
}

// Possible value is something like 1-10
func GetExperienceLevel(percentage float64) string {

	if percentage >= .9 { // avg level ~4
		return `expert`
	}

	if percentage >= .6 { // avg level 3
		return `journeyman`
	}

	if percentage >= .3 { // avg level 2
		return `apprentice`
	}

	if percentage >= .1 { // avg level 1
		return `novice`
	}

	return `scrub`
}
