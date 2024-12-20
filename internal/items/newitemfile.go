package items

import (
	"errors"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/fileloader"
)

func CreateNewItemFile(name string,
	description string,
	value int,
	tp string, stp string,
	damageRoll string,
	uses int,
	keyLockId string,
	questToken string) (int, error) {

	newItemInfo := ItemSpec{
		Name:        name,
		Description: description,
		Value:       value,
		Type:        ItemType(tp),
		Subtype:     ItemSubType(stp),
		Uses:        uses,
		KeyLockId:   keyLockId,
		QuestToken:  questToken,
	}

	if damageRoll != `` {
		dmg := Damage{}
		dmg.InitDiceRoll(damageRoll)
		dmg.DiceRoll = dmg.FormatDiceRoll()
		// Assign it to the item
		newItemInfo.Damage = dmg
	}

	newItemInfo.ItemId = getNextItemId(newItemInfo.Type)
	if newItemInfo.ItemId == 0 {
		return 0, errors.New(`Could not find a new item id to assign.`)
	}

	if err := newItemInfo.Validate(); err != nil {
		return 0, err
	}

	//
	// Before saving, lets zero out the damage info we don't need written.
	// (It can be derrived from the diceroll which is set.)
	//
	newItemInfo.Damage.Attacks = 0
	newItemInfo.Damage.DiceCount = 0
	newItemInfo.Damage.SideCount = 0

	//
	// Save the item
	//
	saveModes := []fileloader.SaveOption{}
	if configs.GetConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}

	if err := fileloader.SaveFlatFile[*ItemSpec](configs.GetConfig().FolderItemData.String(), &newItemInfo, saveModes...); err != nil {
		return 0, err
	}

	// Save to in-memory cache
	items[newItemInfo.Id()] = &newItemInfo

	return newItemInfo.Id(), nil
}

func getNextItemId(t ItemType) int {

	rangeMin := 0
	rangeMax := 0

	tString := string(t)
	for _, iType := range ItemTypes() {

		if iType.Type == tString {
			rangeMin = iType.MinItemId
			rangeMax = iType.MaxItemId
			break
		}

	}

	if rangeMin == 0 && rangeMax == 0 {
		return 0
	}

	lowestFreeId := 1
	for _, iSpec := range items {

		if iSpec.Type != t {
			continue
		}

		if iSpec.ItemId >= lowestFreeId {
			lowestFreeId = iSpec.ItemId + 1
		}
	}

	if lowestFreeId < rangeMin || lowestFreeId > rangeMax {
		return 0
	}

	return lowestFreeId
}
