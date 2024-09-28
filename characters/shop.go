package characters

import (
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/util"
)

type Shop []ShopItem

type ShopItem struct {
	MobId           int `yaml:"mobid,omitempty"`           // Is it a mercenary for sale?
	ItemId          int `yaml:"itemid,omitempty"`          // Is it an item for sale?
	BuffId          int `yaml:"buffid,omitempty"`          // Does this shop keeper apply a buff if purchased?
	Quantity        int `yaml:"quantity,omitempty"`        // How many currently avilable
	QuantityMax     int `yaml:"quantitymax,omitempty"`     // 0 for unlimited, or a maximum that can be stocked at one time
	RestockInterval int `yaml:"restockinterval,omitempty"` // how many rounds between restocks?
	Price           int `yaml:"price,omitempty"`           // If a price is provided, use it

	lastRestockRound uint64 // When was the last time an item was restocked?
}

func (s *Shop) Restock() bool {

	defaultRestockInterval := configs.GetConfig().ShopRestockRounds

	roundNow := util.GetRoundCount()
	restocked := false

	if len(*s) < 1 {
		return restocked
	}

	for i, fsItem := range *s {

		// 0 max means never restocks, always available
		if fsItem.QuantityMax < 1 {
			continue
		}

		if fsItem.Quantity >= fsItem.QuantityMax {
			continue
		}

		itemRestockInterval := uint64(fsItem.RestockInterval)
		if itemRestockInterval == 0 {
			itemRestockInterval = uint64(defaultRestockInterval)
		}

		for roundNow-fsItem.lastRestockRound > itemRestockInterval {
			restocked = true

			fsItem.lastRestockRound += itemRestockInterval
			fsItem.Quantity += 1

			if fsItem.Quantity >= fsItem.QuantityMax {
				fsItem.Quantity = fsItem.QuantityMax
				break
			}
		}

		if restocked {
			fsItem.lastRestockRound = roundNow
		}

		(*s)[i] = fsItem

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
		ItemId:      itemId,
		Quantity:    1,
		QuantityMax: -1,
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
		if (*s)[i].QuantityMax == 0 {
			return true
		}

		(*s)[i].Quantity -= 1

		if (*s)[i].Quantity == 0 && (*s)[i].QuantityMax == -1 {
			(*s) = append((*s)[:i], (*s)[i+1:]...)
		}

		return true

	}

	return false
}

func (s *Shop) GetInstock() Shop {
	ret := Shop{}
	for _, fsItem := range *s {
		if fsItem.Quantity > 0 || fsItem.QuantityMax == 0 {
			ret = append(ret, fsItem)
		}
	}
	return ret
}

func (si *ShopItem) Available() bool {
	return si.Quantity > 0 || si.QuantityMax == 0
}
