package gamelock

import (
	"github.com/GoMudEngine/GoMud/internal/gametime"
	"github.com/GoMudEngine/GoMud/internal/util"
)

const (
	DefaultRelockTime = `1 hour`
)

type Lock struct {
	Difficulty     uint8  `yaml:"difficulty,omitempty"`       // 0 - no lock. greater than zero = difficulty to unlock.
	UnlockedRound  uint64 `yaml:"-"`                          // What round it was unlocked at, when util.GetRoundCount() > UnlockedUntil, it is relocked (set to zero).
	RelockInterval string `yaml:"relockinterval,omitempty"`   // How long until it relocks if unlocked?
	TrapBuffIds    []int  `yaml:"trapbuffids,omitempty,flow"` // if lockpick is failed, a message is displayed about a trap and these are applied.
}

func (l Lock) IsLocked() bool {

	if l.Difficulty == 0 {
		return false
	}

	if l.UnlockedRound == 0 {
		return true
	}

	rndNow := util.GetRoundCount()
	gd := gametime.GetDate(rndNow)

	if l.RelockInterval == `` {
		return rndNow >= gd.AddPeriod(DefaultRelockTime)
	}

	return rndNow >= gd.AddPeriod(l.RelockInterval)
}

func (l *Lock) SetUnlocked() {
	if l.Difficulty > 0 {
		l.UnlockedRound = util.GetRoundCount()
	}
}

func (l *Lock) SetLocked() {
	l.UnlockedRound = 0
}
