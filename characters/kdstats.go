package characters

type KDStats struct {
	TotalKills  int         `json:"totalkills"`  // Quick tally of kills
	Kills       map[int]int `json:"kills"`       // map of MobId to count
	TotalDeaths int         `json:"totaldeaths"` // Quick tally of deaths
}

func (kd *KDStats) GetKDRatio() float64 {
	if kd.TotalDeaths == 0 {
		return float64(kd.TotalKills)
	}
	return float64(kd.TotalKills) / float64(kd.TotalDeaths)
}

func (kd *KDStats) GetKills(mobId ...int) int {
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

func (kd *KDStats) AddKill(mobId int) {
	if kd.Kills == nil {
		kd.Kills = make(map[int]int)
	}
	kd.TotalKills++
	kd.Kills[mobId] = kd.Kills[mobId] + 1
}

func (kd *KDStats) GetDeaths() int {
	return kd.TotalDeaths
}

func (kd *KDStats) AddDeath() {
	kd.TotalDeaths++
}
