package leaderboard

import (
	"math"
	"sync"
	"time"

	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
)

var (

	// Key is type of leaderboard
	leaderboardCache = map[string]Leaderboard{}

	lbLock = sync.RWMutex{}
)

type LeaderboardEntry struct {
	Username       string
	CharacterName  string
	CharacterClass string
	Experience     int
	Level          int
	Gold           int
	Kills          int
}

type Leaderboard []LeaderboardEntry

func Update() map[string]Leaderboard {

	start := time.Now()

	lbLock.Lock()
	defer lbLock.Unlock()

	userCount := 0
	characterCount := 0

	lSize := int(configs.GetStatisticsConfig().LeaderboardSize)

	// Check online users
	for _, u := range users.GetAllActiveUsers() {
		considerUser(u)
	}

	// Check offline users
	users.SearchOfflineUsers(considerUser)

	for lbName, _ := range leaderboardCache {

		if len(leaderboardCache[lbName]) < lSize {

			newEntry := LeaderboardEntry{}

			for i := len(leaderboardCache[lbName]); i < lSize; i++ {
				leaderboardCache[lbName] = append(leaderboardCache[lbName], newEntry)
			}
		}

	}

	lbCopy := map[string]Leaderboard{}
	for name, lbEntries := range leaderboardCache {
		lbCopy[name] = append(Leaderboard{}, lbEntries...)
	}

	mudlog.Info("leaderboard.Update()", "user-processed", userCount, "characters-processed", characterCount, "Time Taken", time.Since(start))

	return lbCopy
}

func Get() map[string]Leaderboard {

	lbLock.RLock()
	lbSize := len(leaderboardCache)
	lbLock.RUnlock()

	if lbSize == 0 {
		return Update()
	}

	lbLock.RLock()
	defer lbLock.RUnlock()

	lbCopy := map[string]Leaderboard{}
	for name, lbEntries := range leaderboardCache {
		lbCopy[name] = append(Leaderboard{}, lbEntries...)
	}

	return lbCopy
}

func considerUser(u *users.UserRecord) bool {

	lSize := int(configs.GetStatisticsConfig().LeaderboardSize)

	allChars := []characters.Character{}
	allChars = append(allChars, *u.Character)
	allChars = append(allChars, characters.LoadAlts(u.UserId)...)

	for _, char := range allChars {

		lbTypes := map[string]int{
			"experience": char.Experience,
			"gold":       char.Gold + char.Bank,
			"kills":      char.KD.TotalKills,
		}

		for lbName, lbValue := range lbTypes {

			if leaderboardCache[lbName] == nil {
				leaderboardCache[lbName] = Leaderboard{}
			}

			lbLowest := math.MaxInt
			if len(leaderboardCache[lbName]) > 0 {

				if lbName == `experience` {
					lbLowest = leaderboardCache[lbName][len(leaderboardCache[lbName])-1].Experience
				} else if lbName == `gold` {
					lbLowest = leaderboardCache[lbName][len(leaderboardCache[lbName])-1].Gold
				} else if lbName == `kills` {
					lbLowest = leaderboardCache[lbName][len(leaderboardCache[lbName])-1].Kills
				}

			}

			if char.Experience == lbLowest && len(leaderboardCache[lbName]) >= lSize {
				return true
			}

			// Add to the list
			addAt := -1
			for i, entry := range leaderboardCache[lbName] {

				if lbName == `experience` {
					if entry.Experience >= lbValue {
						continue
					}
				} else if lbName == `gold` {
					if entry.Gold >= lbValue {
						continue
					}
				} else if lbName == `kills` {
					if entry.Kills >= lbValue {
						continue
					}
				}

				addAt = i
				break
			}

			newEntry := LeaderboardEntry{
				Username:       u.Username,
				CharacterName:  char.Name,
				CharacterClass: skills.GetProfession(char.GetAllSkillRanks()),
				Experience:     char.Experience,
				Level:          char.Level,
				Gold:           char.Gold + char.Bank,
				Kills:          char.KD.TotalKills,
			}

			if addAt == -1 {

				if len(leaderboardCache[lbName]) >= lSize {
					continue
				}

				leaderboardCache[lbName] = append(leaderboardCache[lbName], newEntry)

			} else {

				leaderboardCache[lbName] = append(leaderboardCache[lbName][0:addAt], append(Leaderboard{newEntry}, leaderboardCache[lbName][addAt:]...)...)
				if len(leaderboardCache[lbName]) > lSize {
					leaderboardCache[lbName] = leaderboardCache[lbName][:lSize]
				}

			}

		}

	}

	return true

}
