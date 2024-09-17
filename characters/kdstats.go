package characters

import "strings"

type KDStats struct {
	TotalKills  int            `json:"totalkills"`  // Quick tally of kills
	Kills       map[int]int    `json:"kills"`       // map of MobId to count
	RaceKills   map[string]int `json:"racekills"`   // map of race to count
	ZoneKills   map[string]int `json:"zonekills"`   // map of zone name to count
	TotalDeaths int            `json:"totaldeaths"` // Quick tally of deaths
}

func (kd *KDStats) GetKDRatio() float64 {
	if kd.TotalDeaths == 0 {
		return float64(kd.TotalKills)
	}
	return float64(kd.TotalKills) / float64(kd.TotalDeaths)
}

func (kd *KDStats) GetMobKills(mobId ...int) int {
	if len(mobId) == 0 {
		return kd.TotalKills
	}

	if kd.Kills == nil {
		kd.Kills = make(map[int]int)
	}

	total := 0
	for _, id := range mobId {
		total += kd.Kills[id]
	}
	return total
}

func (kd *KDStats) AddMobKill(mobId int) {
	if kd.Kills == nil {
		kd.Kills = make(map[int]int)
	}
	kd.TotalKills++
	kd.Kills[mobId] = kd.Kills[mobId] + 1
}

func (kd *KDStats) GetRaceKills(race ...string) int {
	if len(race) == 0 {
		return kd.TotalKills
	}

	if kd.RaceKills == nil {
		kd.RaceKills = make(map[string]int)
	}

	total := 0
	for _, raceName := range race {
		raceName = strings.ToLower(raceName)
		total += kd.RaceKills[raceName]
	}
	return total
}

func (kd *KDStats) AddRaceKill(race string) {
	if kd.RaceKills == nil {
		kd.RaceKills = make(map[string]int)
	}

	race = strings.ToLower(race)

	kd.RaceKills[race] = kd.RaceKills[race] + 1
}

func (kd *KDStats) GetZoneKills(zone ...string) int {
	if len(zone) == 0 {
		return kd.TotalKills
	}

	if kd.ZoneKills == nil {
		kd.ZoneKills = make(map[string]int)
	}

	total := 0
	for _, zoneName := range zone {
		zoneName = strings.ToLower(zoneName)
		total += kd.ZoneKills[zoneName]
	}
	return total
}

func (kd *KDStats) AddZoneKill(zoneName string) {
	if kd.ZoneKills == nil {
		kd.ZoneKills = make(map[string]int)
	}

	zoneName = strings.ToLower(zoneName)

	kd.ZoneKills[zoneName] = kd.ZoneKills[zoneName] + 1
}

func (kd *KDStats) GetDeaths() int {
	return kd.TotalDeaths
}

func (kd *KDStats) AddDeath() {
	kd.TotalDeaths++
}
