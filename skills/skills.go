package skills

import (
	"strings"
)

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
	Cast        SkillTag = `cast`        // TODO - far east city?
	DualWield   SkillTag = `dual-wield`  // [LVL 1-4] Fishermans house - ROOM 758
	Map         SkillTag = `map`         // [LVL 1-4] Frostwarden Rangers - ROOM 74
	Enchant     SkillTag = `enchant`     // TODO
	Peep        SkillTag = `peep`        // TODO
	Inspect     SkillTag = `inspect`     // TODO
	Portal      SkillTag = `portal`      // TODO
	Search      SkillTag = `search`      // [LVL 1-4] Frostwarden Rangers - ROOM 74
	Track       SkillTag = `track`       // [LVL 1-4] Frostwarden Rangers - ROOM 74
	Skulduggery SkillTag = `skulduggery` // [LVL 1-4] Thieves Den - ROOM 491
	Brawling    SkillTag = `brawling`    // [LVL 1-4] Soldiers Training Yard - ROOM 829
	Scribe      SkillTag = `scribe`      // [LVL 1-4] Dark Acolyte's Chamber - ROOM 160
	Protection  SkillTag = `protection`  // TODO
	Tame        SkillTag = `tame`        // [LVL 1-4] Give mushroom to fairie in ROOOM 558, train in ROOM 830
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
		"ranger": {
			Search,
			Track,
			Map,
		},
		"monster hunter": {
			Track,
			Tame,
		},
		"sorcerer": {
			Cast,
			Enchant,
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
