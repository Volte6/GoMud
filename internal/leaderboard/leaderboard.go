package leaderboard

import (
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

	updated      = false
	lbGold       = Leaderboard{Name: `Gold`, ValueColor: `experience`}
	lbExperience = Leaderboard{Name: `Experience`, ValueColor: `gold`}
	lbKills      = Leaderboard{Name: `Kills`, ValueColor: `red-bold`}

	lbLock = sync.RWMutex{}
)

type LeaderboardEntry struct {
	UserId         int
	CharacterName  string
	CharacterClass string
	Level          int
	ScoreValue     int
}

type Leaderboard struct {
	Name        string
	ValueColor  string // Numeric 256 color or ansitags alias
	Top         []LeaderboardEntry
	MaxSize     int
	LowestValue int
}

func (l *Leaderboard) Reset(size int) {
	l.MaxSize = size
	l.Top = make([]LeaderboardEntry, l.MaxSize)
	l.LowestValue = 0
}

func (l *Leaderboard) Consider(userId int, char characters.Character, val int) {
	if val == 0 {
		return
	}

	if val < l.LowestValue && l.Top[l.MaxSize-1].UserId != 0 {
		return
	}

	addPosition := -1
	for i := 0; i < l.MaxSize; i++ {

		if l.Top[i].UserId == 0 {
			addPosition = i
			break
		}

		if val > l.Top[i].ScoreValue {
			addPosition = i
			break
		}

	}

	if addPosition > -1 {

		for i := l.MaxSize - 2; i >= addPosition; i-- {
			l.Top[i+1] = l.Top[i]
		}

		// just accept it
		l.Top[addPosition] = LeaderboardEntry{
			UserId:         userId,
			CharacterName:  char.Name,
			CharacterClass: skills.GetProfession(char.GetAllSkillRanks()),
			Level:          char.Level,
			ScoreValue:     val,
		}

		if l.LowestValue == 0 || val < l.LowestValue {
			l.LowestValue = val
		}

	}
}

func Reset(maxSize int) {
	lbGold.Reset(maxSize)
	lbExperience.Reset(maxSize)
	lbKills.Reset(maxSize)
}

func Update() {

	start := time.Now()

	lbLock.Lock()
	defer func() {
		lbLock.Unlock()
	}()

	lbConfig := configs.GetStatisticsConfig().Leaderboards

	Reset(int(lbConfig.Size))

	userCount := 0
	characterCount := 0

	for _, u := range users.GetAllActiveUsers() {

		userCount++
		characterCount++

		if lbConfig.GoldEnabled {
			lbGold.Consider(u.UserId, *u.Character, u.Character.Gold+u.Character.Bank)
		}

		if lbConfig.ExperienceEnabled {
			lbExperience.Consider(u.UserId, *u.Character, u.Character.Experience)
		}

		if lbConfig.KillsEnabled {
			lbKills.Consider(u.UserId, *u.Character, u.Character.KD.TotalKills)
		}

		for _, char := range characters.LoadAlts(u.UserId) {

			characterCount++

			if lbConfig.GoldEnabled {
				lbGold.Consider(u.UserId, char, char.Gold+char.Bank)
			}

			if lbConfig.ExperienceEnabled {
				lbExperience.Consider(u.UserId, char, char.Experience)
			}

			if lbConfig.KillsEnabled {
				lbKills.Consider(u.UserId, char, char.KD.TotalKills)
			}

		}

	}

	// Check offline users
	users.SearchOfflineUsers(func(u *users.UserRecord) bool {

		userCount++
		characterCount++

		if lbConfig.GoldEnabled {
			lbGold.Consider(u.UserId, *u.Character, u.Character.Gold+u.Character.Bank)
		}

		if lbConfig.ExperienceEnabled {
			lbExperience.Consider(u.UserId, *u.Character, u.Character.Experience)
		}

		if lbConfig.KillsEnabled {
			lbKills.Consider(u.UserId, *u.Character, u.Character.KD.TotalKills)
		}

		for _, char := range characters.LoadAlts(u.UserId) {

			characterCount++

			if lbConfig.GoldEnabled {
				lbGold.Consider(u.UserId, char, char.Gold+char.Bank)
			}

			if lbConfig.ExperienceEnabled {
				lbExperience.Consider(u.UserId, char, char.Experience)
			}

			if lbConfig.KillsEnabled {
				lbKills.Consider(u.UserId, char, char.KD.TotalKills)
			}

		}

		return true
	})

	mudlog.Info("leaderboard.Update()", "user-processed", userCount, "characters-processed", characterCount, "Time Taken", time.Since(start))

	updated = true
}

func Get() []Leaderboard {

	lbLock.RLock()

	if !updated {
		lbLock.RUnlock()
		Update()
		lbLock.RLock()
	}

	defer lbLock.RUnlock()

	ret := []Leaderboard{}

	lbConfig := configs.GetStatisticsConfig().Leaderboards

	if lbConfig.GoldEnabled {
		ret = append(ret, lbGold)
	}

	if lbConfig.ExperienceEnabled {
		ret = append(ret, lbExperience)
	}

	if lbConfig.KillsEnabled {
		ret = append(ret, lbKills)
	}

	return ret
}
