package characters

import (
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/util"
)

const (
	StockTemporary = -1
	StockUnlimited = 0
)

type Shop []ShopItem

type ShopItem struct {
	MobId       int    `yaml:"mobid,omitempty"`       // Is it a mercenary for sale?
	ItemId      int    `yaml:"itemid,omitempty"`      // Is it an item for sale?
	BuffId      int    `yaml:"buffid,omitempty"`      // Does this shop keeper apply a buff if purchased?
	PetType     string `yaml:"pettype,omitempty"`     // Does this shop sell pets?
	Quantity    int    `yaml:"quantity,omitempty"`    // How many currently avilable
	QuantityMax int    `yaml:"quantitymax,omitempty"` // 0 for unlimited, or a maximum that can be stocked at one time
	Price       int    `yaml:"price,omitempty"`       // If a price is provided, use it
	RestockRate string `yaml:"restockrate,omitempty"` // 1 day, 1 week, 1 real month, etc

	lastRestockRound uint64 // When was the last time an item was restocked?
}

func (s *Shop) Restock() bool {

	if len(*s) < 1 {
		return false
	}

	defaultRestockRate := configs.GetConfig().ShopRestockRate.String()
	roundNow := util.GetRoundCount()
	restocked := false
	pruneItems := []int{}

	for i, fsItem := range *s {

		restocked = false

		// 0 max means never restocks, always available
		if fsItem.QuantityMax == StockUnlimited {
			continue
		}

		if fsItem.Quantity == fsItem.QuantityMax {
			continue
		}

		itemRestockRate := fsItem.RestockRate
		if itemRestockRate == `` {
			itemRestockRate = defaultRestockRate
		}

		if fsItem.lastRestockRound == 0 {

			if fsItem.QuantityMax != StockUnlimited {
				if fsItem.Quantity < fsItem.QuantityMax {
					fsItem.Quantity = fsItem.QuantityMax
				}
			}

			restocked = true

		} else {
			gd := gametime.GetDate(fsItem.lastRestockRound)

			restockRnd := gd.AddPeriod(itemRestockRate)

			for roundNow >= restockRnd {
				restocked = true

				if fsItem.QuantityMax == StockUnlimited { // unlimited? No adjustment needed
					break
				}

				if fsItem.Quantity == fsItem.QuantityMax { // currently at the max qty? No adjustment needed
					break
				}

				restockRnd = gametime.GetDate(restockRnd).AddPeriod(itemRestockRate)

				// Non Unlimited, Non temporary
				if fsItem.Quantity < fsItem.QuantityMax { // increase stock if needed
					fsItem.Quantity += 1
					continue
				}

				// Temp item handling
				if fsItem.QuantityMax == StockTemporary {

					if fsItem.Quantity > 0 { // decrease stock on temp items
						fsItem.Quantity--
					}

				}

			}
		}

		if restocked {
			fsItem.lastRestockRound = roundNow
		}

		// Once zeroed, prune it
		if fsItem.QuantityMax == StockTemporary && fsItem.Quantity == 0 {
			pruneItems = append(pruneItems, i)
		}

		(*s)[i] = fsItem
	}

	if len(pruneItems) > 0 {
		for pos := len(pruneItems) - 1; pos >= 0; pos-- {
			i := pruneItems[pos]
			(*s) = append((*s)[:i], (*s)[i+1:]...)
		}
	}

	return restocked

}

func (s *Shop) StockItem(itemId int) bool {

	for i, fsItem := range *s {
		if fsItem.ItemId == itemId {
			(*s)[i].Quantity += 1
			return true
		}
	}

	*s = append(*s, ShopItem{
		ItemId:           itemId,
		Quantity:         1,
		QuantityMax:      StockTemporary,
		lastRestockRound: util.GetRoundCount(),
	})

	return true
}

func (s *Shop) Destock(si ShopItem) bool {

	for i, fsItem := range *s {

		if fsItem.ItemId != si.ItemId {
			continue
		}
		if fsItem.MobId != si.MobId {
			continue
		}
		if fsItem.BuffId != si.BuffId {
			continue
		}

		// If unlimited quantity, just return true
		if (*s)[i].QuantityMax == StockUnlimited {
			return true
		}

		(*s)[i].Quantity -= 1

		if (*s)[i].Quantity == 0 && (*s)[i].QuantityMax == StockTemporary {
			(*s) = append((*s)[:i], (*s)[i+1:]...)
		}

		return true

	}

	return false
}

func (s *Shop) GetInstock() Shop {
	ret := Shop{}
	for _, fsItem := range *s {
		if fsItem.Quantity > 0 || fsItem.QuantityMax == StockUnlimited {
			ret = append(ret, fsItem)
		}
	}
	return ret
}

func (si *ShopItem) Available() bool {
	return si.Quantity > 0 || si.QuantityMax == StockUnlimited
}
