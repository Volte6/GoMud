package buffs

import (
	"log/slog"
)

const (
	TriggersLeftExpired = 0 // When it hits this number it will be pruned ASAP
)

type Buff struct {
	BuffId int // Which buff template does it refer to?
	// Need to instance track the following:
	RoundCounter int `yaml:"roundcounter,omitempty"` // How many rounds have passed. Triggers on (RoundCounter%RoundInterval == 0)
	TriggersLeft int `yaml:"triggersleft,omitempty"` // How many times it triggers
}

func (b *Buff) StatMod(statName string) int {
	if b.Expired() {
		return 0
	}
	if buffInfo := GetBuffSpec(b.BuffId); buffInfo != nil {
		if modAmt, ok := buffInfo.StatMods[statName]; ok {
			return modAmt
		}
	}
	return 0
}

func (b *Buff) Expired() bool {
	return b.TriggersLeft <= TriggersLeftExpired
}

// A list of applied buffs
type Buffs struct {
	List      []*Buff
	buffFlags map[Flag][]int // a map of buff flags to the index of the buff
	buffIds   map[int]int    // a map of a buffId to it position in buffList
}

func New() Buffs {
	return Buffs{
		List:      []*Buff{},
		buffFlags: make(map[Flag][]int),
		buffIds:   make(map[int]int),
	}
}

func (bs *Buffs) Validate() {
	if bs.buffFlags == nil {
		bs.buffFlags = make(map[Flag][]int)
	}
	if bs.buffIds == nil {
		bs.buffIds = make(map[int]int)
	}
	if len(bs.List) != len(bs.buffIds) {
		// Rebuild
		bs.buffIds = make(map[int]int)
		bs.buffFlags = make(map[Flag][]int)

		for idx, b := range bs.List {
			bs.buffIds[b.BuffId] = idx
			bSpec := GetBuffSpec(b.BuffId)
			for _, flag := range bSpec.Flags {
				if _, ok := bs.buffFlags[flag]; !ok {
					bs.buffFlags[flag] = []int{}
				}
				bs.buffFlags[flag] = append(bs.buffFlags[flag], idx)
			}
		}
	}
}

func (bs *Buffs) StatMod(statName string) int {
	buffAmt := 0
	for _, b := range bs.List {
		buffAmt += b.StatMod(statName)
	}
	return buffAmt
}

func (bs *Buff) Name() string {
	if sp := GetBuffSpec(bs.BuffId); sp != nil {
		return sp.Name
	}
	return ""
}

func (bs *Buffs) CancelBuffId(buffId int) bool {
	if idx, ok := bs.buffIds[buffId]; ok {
		bs.List[idx].TriggersLeft = TriggersLeftExpired
		return true
	}
	return false
}

func (bs *Buffs) GetBuffIdsWithFlag(action Flag) []int {
	buffIds := []int{}
	for idx := range bs.buffFlags[action] {
		buffIds = append(buffIds, bs.List[idx].BuffId)
	}
	return buffIds
}

func (bs *Buffs) HasFlag(action Flag, expire bool) bool {

	if action != All {
		if _, ok := bs.buffFlags[action]; !ok || len(bs.buffFlags[action]) == 0 {
			return false
		}
	}

	found := false
	for index, b := range bs.List {
		bSpec := GetBuffSpec(b.BuffId)
		for _, p := range bSpec.Flags {
			if b.Expired() {
				continue
			}
			if p == action || action == All {
				found = true

				// If expire is set, need to check the rest of the buffs to possibly expire them too.
				if expire {
					b.TriggersLeft = TriggersLeftExpired
					bs.List[index] = b
					break
				}

				// Otherwise just return found
				return found

			}
		}

	}

	return found
}

func (bs *Buffs) HasBuff(buffId int) bool {
	if _, ok := bs.buffIds[buffId]; ok {
		return true
	}
	return false
}

func (bs *Buffs) RemoveBuff(buffId int) {
	// If this buff is already applied, just increase the trigger count
	if index, ok := bs.buffIds[buffId]; ok {
		bs.List[index].TriggersLeft = TriggersLeftExpired
	}
}

func (bs *Buffs) AddBuff(buffId int) bool {
	if buffInfo := GetBuffSpec(buffId); buffInfo != nil {
		newBuff := Buff{
			BuffId:       buffInfo.BuffId,
			RoundCounter: 0,
			TriggersLeft: buffInfo.TriggerCount,
		}

		if idx, ok := bs.buffIds[buffId]; ok {
			slog.Info("Refreshing Buff", "buffName", buffInfo.Name, "buffId", buffId)
			bs.List[idx].TriggersLeft = buffInfo.TriggerCount

			return true
		}

		bs.List = append(bs.List, &newBuff)
		listIndex := len(bs.List) - 1
		bs.buffIds[buffId] = listIndex
		for _, flag := range buffInfo.Flags {
			if _, ok := bs.buffFlags[flag]; !ok {
				bs.buffFlags[flag] = []int{}
			}
			bs.buffFlags[flag] = append(bs.buffFlags[flag], listIndex)
		}

		slog.Debug("Added Buff", "buffName", buffInfo.Name, "buffId", buffId)
		return true
	}

	return false
}

// Returns what buffs were triggered
func (bs *Buffs) Trigger(buffId ...int) (triggeredBuffs []*Buff) {

	for idx, b := range bs.List {

		// Special case where 1 or more specific buffId's were expectred to trigger (ONLY!)
		// This might happen if a buff needs to trigger before a round begins
		if len(buffId) > 0 {
			for _, id := range buffId {
				if b.BuffId != id {
					continue
				}
			}
		}

		if buffInfo := GetBuffSpec(b.BuffId); buffInfo != nil {

			// If there's no more life left to it, prune it
			// We do this first so that it's the first thing that happens AFTER a full round has already passed.
			if b.TriggersLeft > 0 {
				b.RoundCounter++
				if b.RoundCounter%buffInfo.RoundInterval == 0 {
					// It cannot be pruned unless it is triggered
					triggeredBuffs = append(triggeredBuffs, b)
					b.TriggersLeft--
				}
				bs.List[idx] = b
			}

		}

	}

	return triggeredBuffs
}

func (bs *Buffs) GetAllBuffs(buffId ...int) []*Buff {
	retBuffs := []*Buff{}
	for _, b := range bs.List {
		if !b.Expired() {

			if len(buffId) > 0 {
				for _, id := range buffId {
					if b.BuffId != id {
						continue
					}
					retBuffs = append(retBuffs, b)
				}
			} else {
				retBuffs = append(retBuffs, b)
			}

		}
	}
	return retBuffs
}

func (bs *Buffs) Prune() (prunedBuffs []*Buff) {

	if len(bs.List) == 0 {
		return prunedBuffs
	}

	var prune bool = false
	for i := len(bs.List) - 1; i >= 0; i-- {

		prune = false

		b := bs.List[i] // Get a ptr to the data within the slice

		buffInfo := GetBuffSpec(b.BuffId)

		if buffInfo == nil {
			prune = true
		} else {
			// If there's no more life left to it, prune it
			// We do this first so that it's the first thing that happens AFTER a full round has already passed.
			if b.Expired() {
				prune = true
			}
		}

		if prune {
			prunedBuffs = append(prunedBuffs, b)
			// remove the buff
			bs.List = append(bs.List[:i], bs.List[i+1:]...)
		}
	}

	// Since pruning occured, rebuild the lookups
	if prune {
		bs.buffFlags = make(map[Flag][]int)
		bs.buffIds = make(map[int]int)
		for idx, b := range bs.List {

			bs.buffIds[b.BuffId] = idx
			for _, flag := range GetBuffSpec(b.BuffId).Flags {
				if _, ok := bs.buffFlags[flag]; !ok {
					bs.buffFlags[flag] = []int{}
				}
				bs.buffFlags[flag] = append(bs.buffFlags[flag], idx)
			}
		}
	}

	return prunedBuffs
}
