package characters

import "github.com/volte6/mud/items"

type Worn struct {
	Weapon  items.Item `yaml:"weapon,omitempty"`
	Offhand items.Item `yaml:"offhand,omitempty"`
	Head    items.Item `yaml:"head,omitempty"`
	Neck    items.Item `yaml:"neck,omitempty"`
	Body    items.Item `yaml:"body,omitempty"`
	Belt    items.Item `yaml:"belt,omitempty"`
	Gloves  items.Item `yaml:"gloves,omitempty"`
	Ring    items.Item `yaml:"ring,omitempty"`
	Legs    items.Item `yaml:"legs,omitempty"`
	Feet    items.Item `yaml:"feet,omitempty"`
}

func (w *Worn) StatMod(stat ...string) int {

	return w.Weapon.StatMod(stat...) +
		w.Offhand.StatMod(stat...) +
		w.Head.StatMod(stat...) +
		w.Neck.StatMod(stat...) +
		w.Body.StatMod(stat...) +
		w.Belt.StatMod(stat...) +
		w.Gloves.StatMod(stat...) +
		w.Ring.StatMod(stat...) +
		w.Legs.StatMod(stat...) +
		w.Feet.StatMod(stat...)
}

func (w *Worn) EnableAll() {
	if w.Weapon.ItemId < 0 {
		w.Weapon = items.Item{}
	}
	if w.Offhand.ItemId < 0 {
		w.Offhand = items.Item{}
	}
	if w.Head.ItemId < 0 {
		w.Head = items.Item{}
	}
	if w.Neck.ItemId < 0 {
		w.Neck = items.Item{}
	}
	if w.Body.ItemId < 0 {
		w.Body = items.Item{}
	}
	if w.Belt.ItemId < 0 {
		w.Belt = items.Item{}
	}
	if w.Gloves.ItemId < 0 {
		w.Gloves = items.Item{}
	}
	if w.Ring.ItemId < 0 {
		w.Ring = items.Item{}
	}
	if w.Legs.ItemId < 0 {
		w.Legs = items.Item{}
	}
	if w.Feet.ItemId < 0 {
		w.Feet = items.Item{}
	}
}

func (w *Worn) GetAllItems() []items.Item {
	iList := []items.Item{}
	if w.Weapon.ItemId > 0 {
		iList = append(iList, w.Weapon)
	}
	if w.Offhand.ItemId > 0 {
		iList = append(iList, w.Offhand)
	}
	if w.Head.ItemId > 0 {
		iList = append(iList, w.Head)
	}
	if w.Neck.ItemId > 0 {
		iList = append(iList, w.Neck)
	}
	if w.Body.ItemId > 0 {
		iList = append(iList, w.Body)
	}
	if w.Belt.ItemId > 0 {
		iList = append(iList, w.Belt)
	}
	if w.Gloves.ItemId > 0 {
		iList = append(iList, w.Gloves)
	}
	if w.Ring.ItemId > 0 {
		iList = append(iList, w.Ring)
	}
	if w.Legs.ItemId > 0 {
		iList = append(iList, w.Legs)
	}
	if w.Feet.ItemId > 0 {
		iList = append(iList, w.Feet)
	}
	return iList
}
