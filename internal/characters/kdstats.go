package characters

import "fmt"

type KDStats struct {
	TotalKills  int         `json:"totalkills,omitempty"`  // Quick tally of kills
	Kills       map[int]int `json:"kills,omitempty"`       // map of MobId to count
	TotalDeaths int         `json:"totaldeaths,omitempty"` // Quick tally of deaths

	TotalPvpKills  int            `json:"totalpvpkills,omitempty"`  // Quick tally of pvp kills
	PlayerKills    map[string]int `json:"playerkills,omitempty"`    // map of userid:username to count
	PlayerDeaths   map[string]int `json:"playerdeaths,omitempty"`   // map of userid:username to count
	TotalPvpDeaths int            `json:"totalpvpdeaths,omitempty"` // Quick tally of pvp deaths
}

func (kd *KDStats) GetMobKDRatio() float64 {
	if kd.TotalDeaths == 0 {
		return float64(kd.TotalKills)
	}
	return float64(kd.TotalKills) / float64(kd.TotalDeaths)
}

func (kd *KDStats) GetPvpKDRatio() float64 {
	if kd.TotalPvpDeaths == 0 {
		return float64(kd.TotalPvpKills)
	}
	return float64(kd.TotalPvpKills) / float64(kd.TotalPvpDeaths)
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

func (kd *KDStats) AddPlayerKill(killedUserId int, killedCharName string) {
	if kd.PlayerKills == nil {
		kd.PlayerKills = make(map[string]int)
	}

	keyName := fmt.Sprintf(`%d:%s`, killedUserId, killedCharName)

	kd.TotalPvpKills++
	kd.PlayerKills[keyName] = kd.PlayerKills[keyName] + 1
}

func (kd *KDStats) AddPlayerDeath(killedByUserId int, killedByCharName string) {
	if kd.PlayerDeaths == nil {
		kd.PlayerDeaths = make(map[string]int)
	}

	keyName := fmt.Sprintf(`%d:%s`, killedByUserId, killedByCharName)
	kd.PlayerDeaths[keyName] = kd.PlayerDeaths[keyName] + 1
}

func (kd *KDStats) AddMobKill(mobId int) {
	if kd.Kills == nil {
		kd.Kills = make(map[int]int)
	}
	kd.TotalKills++
	kd.Kills[mobId] = kd.Kills[mobId] + 1
}

func (kd *KDStats) GetMobDeaths() int {
	return kd.TotalDeaths
}

func (kd *KDStats) GetPvpDeaths() int {
	return kd.TotalPvpDeaths
}

func (kd *KDStats) AddMobDeath() {
	kd.TotalDeaths++
}

func (kd *KDStats) AddPvpDeath() {
	kd.TotalPvpDeaths++
}
