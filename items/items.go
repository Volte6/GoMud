package items

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/volte6/mud/util"
)

//
// Item is used for item instances
// Flat specs are found by loading the spec of the item id.
// Anything in this struct is mutable.
//

var (
	ItemDisabledSlot = Item{ItemId: -1}
)

// Instance properties that may change
type Item struct {
	ItemId        int            `yaml:"itemid,omitempty"`
	Blob          string         `yaml:"blob,omitempty"`          // Does this item have a blob? Should be base64 encoded.
	Uses          int            `yaml:"uses,omitempty"`          // How many uses it has left
	LastUsedRound uint64         `yaml:"lastusedround,omitempty"` // Last round this item was used
	Spec          *ItemSpec      `yaml:"overrides,omitempty"`
	Uncursed      bool           `yaml:"uncursed,omitempty"`     // Is this item uncursed?
	Enchantments  uint8          `yaml:"enchantments,omitempty"` // Is this item enchanted?
	tempDataStore map[string]any // Temporary data store for this item. Not saved to disk.
}

func New(itemId int) Item {
	itemSpec := GetItemSpec(itemId)

	newItm := Item{}
	if itemSpec != nil {
		newItm.ItemId = itemId
		if itemSpec.Uses > 0 {
			newItm.Uses = itemSpec.Uses
		}
	}

	return newItm
}

func (i *Item) GetScript() string {
	return i.GetSpec().GetScript()
}

// performs a break test and returns true if the item breaks
// Pass a uint8 to increase the chance of breaking.
func (i *Item) BreakTest(increaseChance ...int) bool {
	bc := i.GetSpec().BreakChance
	if bc < 1 {
		return false
	}
	randNum := uint8(util.Rand(100))
	if len(increaseChance) > 0 {
		if uint8(increaseChance[0]) >= randNum {
			randNum = 0
		} else {
			randNum -= uint8(increaseChance[0])
		}
	}
	return bc > randNum
}

func (i *Item) SetTempData(key string, value any) {

	if i.tempDataStore == nil {
		i.tempDataStore = make(map[string]any)
	}

	if value == nil {
		delete(i.tempDataStore, key)
		return
	}
	i.tempDataStore[key] = value
}

func (i *Item) GetTempData(key string) any {

	if i.tempDataStore == nil {
		i.tempDataStore = make(map[string]any)
	}

	if value, ok := i.tempDataStore[key]; ok {
		return value
	}
	return nil
}

func (i Item) IsDisabled() bool {
	return i.ItemId < 0
}

func (i *Item) Validate() {
	if i.ItemId < 1 {
		return
	}

	iSpec := i.GetSpec()
	if iSpec.ItemId > 0 {
		if i.Uses == 0 && iSpec.Uses > 0 {
			i.Uses = iSpec.Uses
		}
	}

}

func (i *Item) GetLongDescription() string {

	iSpec := i.GetSpec()

	longDesc := strings.Builder{}

	longDesc.WriteString(iSpec.Description)

	if iSpec.Type == Readable {

		longDesc.WriteString("\n")
		longDesc.WriteString(` - You should probably <ansi fg="command">read</ansi> this.`)

	} else if iSpec.Subtype == Drinkable {

		longDesc.WriteString("\n")
		longDesc.WriteString(` - You could probably <ansi fg="command">drink</ansi> this.`)

	} else if iSpec.Subtype == Edible {

		longDesc.WriteString("\n")
		longDesc.WriteString(` - You could probably <ansi fg="command">eat</ansi> this.`)

	} else if iSpec.Type == Lockpicks {

		longDesc.WriteString("\n")
		longDesc.WriteString(` - These are used with the <ansi fg="command">picklock</ansi> command.`)

	} else if iSpec.Type == Key {

		longDesc.WriteString("\n")
		longDesc.WriteString(` - When you find the right door, keys are added to your <ansi fg="command">keyring</ansi> automatically.`)

	} else if iSpec.Subtype == Wearable {

		longDesc.WriteString("\n")
		longDesc.WriteString(fmt.Sprintf(`- It looks like wearable %s equipment.`, iSpec.Type))

	} else if iSpec.Type == Weapon {

		longDesc.WriteString("\n")
		longDesc.WriteString(fmt.Sprintf(`- It looks like a %d-Handed weapon.`, iSpec.Hands))

		if iSpec.Subtype == Claws {

			longDesc.WriteString("\n")
			longDesc.WriteString(`- It looks like a claws weapon. These can be dual wielded without training.`)

		} else if iSpec.Subtype == Shooting {

			longDesc.WriteString("\n")
			longDesc.WriteString(`- This can fired into adjacent areas. (<ansi fg="command">help shoot</ansi>)`)

		}

		if iSpec.WaitRounds > 0 {

			longDesc.WriteString("\n")
			longDesc.WriteString(fmt.Sprintf(`- It requires an extra %d round(s) between attacks.`, iSpec.WaitRounds))

		}

	}

	return longDesc.String()
}

func (i *Item) IsBetterThan(otherItm Item) bool {

	if otherItm.ItemId < 1 {
		return i.ItemId > 0 // As long as the other item isn't also zero, it's better.
	}
	// Whichever is higher value is better
	return i.GetSpec().Value > otherItm.GetSpec().Value
}

func (i *Item) GetSpec() ItemSpec {
	if i.Spec != nil {
		return *i.Spec
	}
	iSpec := GetItemSpec(i.ItemId)
	if iSpec == nil {
		iSpec = &ItemSpec{}
	}
	return *iSpec
}

func (i *Item) Rename(newName string) {
	if i.Spec == nil {
		specCopy := *GetItemSpec(i.ItemId)
		specCopy.Name = newName
		specCopy.DisplayName = ``
		i.Spec = &specCopy
	}
}

func (i *Item) IsEnchanted() bool {
	return i.Enchantments > 0
}

func (i *Item) UnEnchant() {
	if i.IsEnchanted() {
		i.Spec = nil
		i.Enchantments = 0
	}
}

// enchantmentLevel is 0-100. If 0(zero) remove any enchantments.
func (i *Item) Enchant(damageBonus int, defenseBonus int, statBonus map[string]int, cursed bool) {

	var newSpec ItemSpec

	if i.Spec == nil {
		specCopy := *GetItemSpec(i.ItemId)
		newSpec = specCopy
	} else {
		newSpec = *i.Spec
	}

	newSpec.Damage.BonusDamage += damageBonus
	newSpec.DamageReduction += defenseBonus

	if newSpec.StatMods == nil {
		newSpec.StatMods = make(map[string]int)
	}

	for statName, statBonusAmt := range statBonus {
		if _, ok := newSpec.StatMods[statName]; !ok {
			newSpec.StatMods[statName] = 0
		}
		newSpec.StatMods[statName] += statBonusAmt
	}

	i.Enchantments++

	newSpec.Cursed = cursed

	newSpec.Damage.FormatDiceRoll()
	newSpec.AutoCalculateValue()

	i.Spec = &newSpec
}

func (i *Item) Uncurse() {
	i.Uncursed = true
}

func (i *Item) IsCursed() bool {
	return i.GetSpec().Cursed && !i.Uncursed
}

// Gets the specifics of the item damage
// Considers overrides
func (i *Item) GetDiceRoll() (attacks int, dCount int, dSides int, bonus int, buffOnCrit []int) {
	if i.ItemId < 1 {
		return 1, 1, 3, 0, []int{} // Default Damages
	}
	dmg := i.GetDamage()
	return dmg.Attacks, dmg.DiceCount, dmg.SideCount, dmg.BonusDamage, dmg.CritBuffIds
}

func (i *Item) IsSpecial() bool {
	iSpec := i.GetSpec()
	if len(i.Blob) > 0 {
		return true
	}
	if iSpec.Uses > 0 && iSpec.Uses != i.Uses {
		return true
	}
	if i.Spec != nil {
		return true
	}

	return false
}

func (i *Item) GetDamage() Damage {
	return i.GetSpec().Damage
}

// Returns a random number up to the total possible reduction for this item.
func (i *Item) GetDefense() int {
	itemInfo := i.GetSpec()
	return itemInfo.DamageReduction
}

func (i *Item) Equals(b Item) bool {

	if i.ItemId != b.ItemId {
		return false
	}

	if i.Blob != b.Blob {
		return false
	}

	if i.Uses != b.Uses {
		return false
	}

	if i.Spec != b.Spec {
		return false
	}

	// If there is a spec defined on this item, then the other item should also have a spec defined pointing to the same address.
	if i.Spec != nil && i.Spec != b.Spec {
		return false
	}

	return true
}

func (i *Item) IsValid() bool {

	if itemInfo := GetItemSpec(i.ItemId); itemInfo != nil {
		return true
	}
	return false
}

func (i *Item) GetBlob() string {
	if len(i.Blob) == 0 {
		return ``
	}

	decoded := util.Decode(i.Blob)
	return string(util.Decompress(decoded))
}

func (i *Item) SetBlob(blob string) {
	compressed := util.Compress([]byte(blob))
	i.Blob = util.Encode(compressed)
}

func (i *Item) AttrString() string {

	flags := []string{}

	if i.IsCursed() {
		flags = append(flags, `<ansi fg="item-cursed">c</ansi>`)
	}
	if i.IsEnchanted() {
		flags = append(flags, `<ansi fg="item-enchanted">e</ansi>`)
	}

	if len(flags) == 0 {
		return ``
	}

	return fmt.Sprintf(`<ansi fg="item-flags">[%s]</ansi>`, strings.Join(flags, ``))
}

func (i *Item) DisplayName() string {
	if i.ItemId < 1 { // Used to represent item slots that are disabled
		if i.ItemId == 0 { // Used to represent item slots that are empty
			return `<ansi fg="item-nothing">-nothing-</ansi>`
		} else {
			return `<ansi fg="item-nothing">***disabled***</ansi>`
		}
	}

	spec := i.GetSpec()
	if spec.DisplayName != `` {
		return spec.DisplayName
	}
	return spec.Name
}

func (i *Item) Name() string {

	if i.ItemId < 1 { // Used to represent item slots that are disabled
		if i.ItemId == 0 { // Used to represent item slots that are empty
			return `-nothing-`
		} else {
			return `***disabled***`
		}
	}

	return i.GetSpec().Name
}

func (i *Item) NameSimple() string {

	if i.ItemId < 1 { // Used to represent item slots that are disabled
		if i.ItemId == 0 { // Used to represent item slots that are empty
			return `-nothing-`
		} else {
			return `***disabled***`
		}
	}

	return i.GetSpec().NameSimple
}

func (i *Item) NameComplex() string {

	if i.ItemId < 1 { // Used to represent item slots that are disabled
		if i.ItemId == 0 { // Used to represent item slots that are empty
			return `<ansi fg="item-nothing">-nothing-</ansi>`
		} else {
			return `<ansi fg="item-nothing">***disabled***</ansi>`
		}
	}

	nm := i.DisplayName()

	if i.GetSpec().Damage.BonusDamage > 0 {
		nm = fmt.Sprintf(`%s <ansi fg="item-bonus-damage">+%d</ansi>`, nm, i.GetSpec().Damage.BonusDamage)
	}
	flagsStr := i.AttrString()
	if flagsStr != `` {
		nm = fmt.Sprintf(`%s %s`, flagsStr, nm)
	}
	return nm
}

func (i *Item) NameMatch(input string, allowContains bool) (partialMatch bool, fullMatch bool) {

	if i.ItemId < 1 { // Used to represent item slots that are empty
		return false, false
	}

	input = strings.ToLower(input)
	simpleName := strings.ToLower(i.Name())

	if allowContains {
		if strings.Contains(simpleName, input) {
			if simpleName == input {
				return true, true
			}
			return true, false
		}
	}

	if strings.HasPrefix(simpleName, input) {
		if simpleName == input {
			return true, true
		}
		return true, false
	}

	return false, false
}

func (i *Item) StatMod(statName ...string) int {

	if i.ItemId < 1 {
		return 0
	}

	retAmt := 0

	itemInfo := i.GetSpec()
	if len(itemInfo.StatMods) == 0 {
		return retAmt
	}

	for _, stat := range statName {
		if modAmt, ok := itemInfo.StatMods[stat]; ok {
			retAmt += modAmt
		}
	}

	return retAmt
}

func startsWithVowel(s string) bool {
	if len(s) == 0 {
		return false
	}

	firstChar := unicode.ToLower(rune(s[0]))
	return firstChar == 'a' || firstChar == 'e' || firstChar == 'i' || firstChar == 'o' || firstChar == 'u'
}

// Provided a name and a list of items, find the first item that matches the name
// Will first provide a pair of starts-width and exact matches,
// and if not found then a contains.
func FindMatchIn(itemName string, items ...Item) (pMatch Item, fMatch Item) {

	if len(itemName) > 1 {
		if itemName[0] == '!' { // Special meaning to specify an item
			itemIdMatch, _ := strconv.Atoi(itemName[1:])

			for _, itm := range items {

				if itemIdMatch > 0 {
					if itm.ItemId != itemIdMatch {
						continue
					}
					return itm, itm
				}
			}
			return Item{}, Item{}
		}
	}

	itemName, itemNumber := util.GetMatchNumber(itemName)

	var matchItem Item
	var closeMatchItem Item

	var matchItemCt int = 0
	var closeMatchItemCt int = 0

	for _, i := range items {

		part, full := i.NameMatch(itemName, false)

		if part {
			closeMatchItemCt++
			if closeMatchItemCt == itemNumber {
				closeMatchItem = i
			}
		}

		if full {
			matchItemCt++
			if matchItemCt == itemNumber {
				matchItem = i
				break
			}
		}

	}

	// If no "starts with" or "exact" matches are found, try and find the first items that contain the supplied name
	// Note: Can't have an exact match if there was never a close match
	if closeMatchItem.ItemId == 0 {
		closeMatchItemCt = 0
		for _, i := range items {
			part, _ := i.NameMatch(itemName, true)

			if part {
				closeMatchItemCt++
				if closeMatchItemCt == itemNumber {
					closeMatchItem = i
					break
				}
			}

		}

	}

	if matchItem.ItemId > 0 {
		return Item{}, matchItem
	}

	if closeMatchItem.ItemId > 0 {
		return closeMatchItem, Item{}
	}

	return Item{}, Item{}
}
