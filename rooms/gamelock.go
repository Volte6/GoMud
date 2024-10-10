package rooms

import (
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/util"
)

type GameLock struct {
	Difficulty    uint8  `yaml:"difficulty,omitempty"` // 0 - no lock. greater than zero = difficulty to unlock.
	UnlockedUntil uint64 `yaml:"-"`                    // What round it was unlocked at, when util.GetRoundCount() > UnlockedUntil, it is relocked (set to zero).
}

func (l GameLock) IsLocked() bool {
	return l.Difficulty > 0 && l.UnlockedUntil < util.GetRoundCount()
}

func (l *GameLock) SetUnlocked() {
	if l.Difficulty > 0 && l.UnlockedUntil < util.GetRoundCount() {
		l.UnlockedUntil = util.GetRoundCount() + uint64(configs.GetConfig().MinutesToRounds(5))
	}
}

func (l *GameLock) SetLocked() {
	l.UnlockedUntil = 0
}
