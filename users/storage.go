package users

import "github.com/volte6/mud/items"

type Storage struct {
	Items []items.Item
}

func (s *Storage) GetItems() []items.Item {
	return append([]items.Item{}, s.Items...)
}

func (s *Storage) FindItem(itemName string) (items.Item, bool) {

	if itemName == `` {
		return items.Item{}, false
	}

	closeMatchItem, matchItem := items.FindMatchIn(itemName, s.Items...)

	if matchItem.ItemId != 0 {
		return matchItem, true
	}

	if closeMatchItem.ItemId != 0 {
		return closeMatchItem, true
	}

	return items.Item{}, false
}

func (s *Storage) AddItem(i items.Item) bool {
	if i.ItemId < 1 {
		return false
	}
	s.Items = append(s.Items, i)
	return true
}

func (s *Storage) RemoveItem(i items.Item) bool {
	for j := len(s.Items) - 1; j >= 0; j-- {
		if s.Items[j].Equals(i) {
			s.Items = append(s.Items[:j], s.Items[j+1:]...)
			return true
		}
	}
	return false
}
