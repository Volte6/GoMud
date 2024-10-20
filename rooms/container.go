package rooms

import "github.com/volte6/gomud/items"

type Container struct {
	Lock  GameLock     `yaml:"lock,omitempty"` // 0 - no lock. greater than zero = difficulty to unlock.
	Items []items.Item `yaml:"-"`              // Don't save contents, let room prep handle. What is found inside of the chest.
	Gold  int          `yaml:"-"`              // Don't save contents, let room prep handle. How much gold is found inside of the chest.
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
