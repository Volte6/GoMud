package rooms

import "github.com/volte6/gomud/internal/items"

type Container struct {
	Lock         GameLock     `yaml:"lock,omitempty"`         // 0 - no lock. greater than zero = difficulty to unlock.
	Items        []items.Item `yaml:"items,omitempty"`        // Save contents now, since players can put new items in there
	Gold         int          `yaml:"gold,omitempty"`         // Save contents now, since players can put new items in there
	DespawnRound uint64       `yaml:"despawnround,omitempty"` // If this is set, it's a chest that will disappear with time.
}

func (c Container) HasLock() bool {
	return c.Lock.Difficulty > 0
}

func (c *Container) AddItem(i items.Item) {
	c.Items = append(c.Items, i)
}

func (c *Container) RemoveItem(i items.Item) {
	for j := len(c.Items) - 1; j >= 0; j-- {
		if c.Items[j].Equals(i) {
			c.Items = append(c.Items[:j], c.Items[j+1:]...)
			break
		}
	}
}

func (c *Container) FindItem(itemName string) (items.Item, bool) {

	// Search floor
	closeMatchItem, matchItem := items.FindMatchIn(itemName, c.Items...)

	if matchItem.ItemId != 0 {
		return matchItem, true
	}

	if closeMatchItem.ItemId != 0 {
		return closeMatchItem, true
	}

	return items.Item{}, false
}
