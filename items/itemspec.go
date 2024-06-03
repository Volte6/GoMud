package items

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"time"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/fileloader"
	"github.com/volte6/mud/util"
)

type ItemType string
type ItemSubType string
type Element string
type Intensity string
type TokenName string

type WeaponHands = int

var (
	items map[int]*ItemSpec = make(map[int]*ItemSpec)
)

const (
	Unknown ItemType = ""

	// Equipment
	Weapon  ItemType = "weapon"
	Offhand ItemType = "offhand"
	Head    ItemType = "head"
	Neck    ItemType = "neck"
	Body    ItemType = "body"
	Belt    ItemType = "belt"
	Gloves  ItemType = "gloves"
	Ring    ItemType = "ring"
	Legs    ItemType = "legs"
	Feet    ItemType = "feet"
	// Consumables
	Potion  ItemType = "potion"
	Food    ItemType = "food"
	Drink   ItemType = "drink"
	Scroll  ItemType = "scroll"
	Grenade ItemType = "grenade" // Expected to be thrown

	// Other
	Container ItemType = "container"
	Readable  ItemType = "readable"  // Something with writing to reveal when read
	Currency  ItemType = "currency"  // it's gold, basically.
	Key       ItemType = "key"       // A key for a door
	Object    ItemType = "object"    // A mundane object
	Gemstone  ItemType = "gemstone"  // A gem
	Lockpicks ItemType = "lockpicks" // Used for lockpicking

	// Subtypes for wearables
	Wearable  ItemSubType = "wearable"
	Drinkable ItemSubType = "drinkable"
	Edible    ItemSubType = "edible"
	Usable    ItemSubType = "usable"
	Throwable ItemSubType = "throwable" // If dropped/thrown, triggers buff effects on room and is lost
	Mundane   ItemSubType = "mundane"

	// Subtypes for weapons, chooses attack messages.
	Generic     ItemSubType = "generic"
	Bludgeoning ItemSubType = "bludgeoning"
	Cleaving    ItemSubType = "cleaving"
	Stabbing    ItemSubType = "stabbing"
	Slashing    ItemSubType = "slashing"
	Shooting    ItemSubType = "shooting" // bows, crossbows, guns, etc.
	Claws       ItemSubType = "claws"

	BlobContent ItemSubType = "blobcontent"
	Gold        ItemSubType = "gold"

	OneHanded WeaponHands = 1
	TwoHanded WeaponHands = 2

	Fire        Element = "fire"
	Water       Element = "water"
	Ice         Element = "ice"
	Electricity Element = "electricity"
	Acid        Element = "acid"
	Life        Element = "life"
	Death       Element = "death"

	// Intensity of the attack
	Prepare  Intensity = "prepare"
	Wait     Intensity = "wait"
	Miss     Intensity = "miss"
	Weak     Intensity = "weak"
	Normal   Intensity = "normal"
	Heavy    Intensity = "heavy"
	Critical Intensity = "critical"

	// Tokens
	TokenItemName     TokenName = "{itemname}"
	TokenSource       TokenName = "{source}"
	TokenSourceType   TokenName = "{sourcetype}" // will be 'user' or 'mob'
	TokenTarget       TokenName = "{target}"
	TokenTargetType   TokenName = "{targettype}" // will be 'user' or 'mob'
	TokenUsesLeft     TokenName = "{usesleft}"
	TokenDamage       TokenName = "{damage}"
	TokenEntranceName TokenName = "{entrancename}"
	TokenExitName     TokenName = "{exitname}"

	POVUser  = 0
	POVOther = 1

	itemDataFilesFolderPath = "_datafiles/items"
)

type Damage struct {
	Attacks     int    `yaml:"attacks,omitempty"` // How many attacks this weapon gets (usually 1)
	DiceRoll    string // 1d6, etc.
	CritBuffIds []int  `yaml:"critbuffids,omitempty"` // If this damage is a crit, what buffs does it apply?
	DiceCount   int    // how many dice to roll for this weapons damage
	SideCount   int    // how many sides per dice roll
	BonusDamage int    `yaml:"bonusdamage,omitempty"` // flat damage bonus, so for example 1d6+1
}

type ItemMessage string

// Attack messages
type AttackMessageOptions []ItemMessage
type AttackEffects map[Intensity]AttackMessageOptions
type AttackMessages map[ItemSubType]AttackEffects

// The blueprint for an item
type ItemSpec struct {
	ItemId          int
	Value           int
	Uses            int         `yaml:"uses,omitempty"`            // How many uses it starts with
	BuffIds         []int       `yaml:"buffids,omitempty"`         // What buffs it can apply (if used correctly)
	DamageReduction int         `yaml:"damagereduction,omitempty"` // % of damage it reduces when it blocks attacks
	WaitRounds      int         `yaml:"waitrounds,omitempty"`      // How many extra rounds each combat requires
	Hands           WeaponHands `yaml:"hands"`                     // How many hands it takes to wield
	Name            string
	NameSimple      string // A simpler name for the item, for example "Golden Battleaxe" should be "Battleaxe" or "Axe" for simple
	Description     string
	QuestToken      string `yaml:"questtoken,omitempty"` // Grants this quest if given/picked up
	Type            ItemType
	Subtype         ItemSubType
	Damage          Damage
	Element         Element
	StatMods        map[string]int `yaml:"statmods,omitempty"`  // What stats it modifies when equipped
	Cursed          bool           `yaml:"cursed,omitempty"`    // Can't be removed once equipped
	KeyLockId       string         `yaml:"keylockid,omitempty"` // Example: `778-north` - If it's a key, what lock does it open? roomid-exitname etc.
}

func (i Element) String() string {
	return string(i)
}

func (i ItemType) String() string {
	return string(i)
}

func (i ItemSubType) String() string {
	return string(i)
}

func (d *Damage) String() string {
	if d.DiceRoll == "" {
		return "N/A"
	}
	return d.DiceRoll
}

func (d *Damage) FormatDiceRoll() string {

	d.DiceRoll = util.FormatDiceRoll(d.Attacks, d.DiceCount, d.SideCount, d.BonusDamage, d.CritBuffIds)

	return d.DiceRoll
}

func (d *Damage) InitDiceRoll(dRoll string) {
	// If diceroll is specified, it overrides whatever stats are already there
	if len(dRoll) < 1 {
		return
	}

	d.Attacks, d.DiceCount, d.SideCount, d.BonusDamage, _ = util.ParseDiceRoll(dRoll)
}

func FindItemByName(name string) int {
	name = strings.ToLower(name)

	for _, item := range items {
		if strings.ToLower(item.Name) == name {
			return item.ItemId
		}
	}

	for _, item := range items {
		if strings.HasPrefix(strings.ToLower(item.Name), name) {
			return item.ItemId
		}
	}

	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Name), name) {
			return item.ItemId
		}
	}

	return 0
}

func GetAllItemNames() []string {

	itemNames := []string{}
	for _, item := range items {
		itemNames = append(itemNames, item.Name)
	}
	return itemNames
}

// Presumably to ensure the datafile hasn't messed something up.
func (i *ItemSpec) Id() int {
	return i.ItemId
}

func CanBackstab(iSubType ItemSubType) bool {
	if iSubType == Cleaving || iSubType == Stabbing || iSubType == Slashing || iSubType == Claws {
		return true
	}
	return false
}

func (i *ItemSpec) AutoCalculateValue() {

	val := 5 // base value of 5

	// Weapon based damage valuation
	val += (i.Damage.DiceCount * i.Damage.DiceCount) * (i.Damage.SideCount * i.Damage.SideCount * 2)
	val += i.Damage.BonusDamage * 25
	// Armor based damage valuation
	val += (i.DamageReduction * i.DamageReduction) * 17

	// Get the value of any buff it applies
	for _, buffId := range i.BuffIds {
		if buffSpec := buffs.GetBuffSpec(buffId); buffSpec != nil {
			val += buffSpec.GetValue()
		}
	}

	for _, statMod := range i.StatMods {
		val += statMod * 11
	}

	// Special considerations
	if i.Uses > 1 {
		val *= i.Uses
	}

	if i.Type == Lockpicks {
		val *= 2
	}

	if i.Hands > 1 {
		val = int(math.Ceil(float64(val) * 1.25))
	}

	if i.Type == Ring {
		// rings are atomatically worth more, since they are jewelry
		val *= 2
	}

	i.Value = val
}

// Presumably to ensure the datafile hasn't messed something up.
func (i *ItemSpec) Validate() error {

	if i.Type == Weapon {
		if i.Hands == 0 {
			i.Hands = 1
		}
		if i.Damage.Attacks < 1 {
			i.Damage.Attacks = 1
		}
	}

	if i.NameSimple == `` {
		i.NameSimple = i.Name
	}

	i.Damage.InitDiceRoll(i.Damage.DiceRoll)
	i.Damage.FormatDiceRoll()

	if i.Value < 1 {
		i.AutoCalculateValue()
	}

	return nil
}

func (i *ItemSpec) Filename() string {
	return fmt.Sprintf("%d.yaml", i.ItemId)
}

func (i *ItemSpec) Filepath() string {
	if i.ItemId >= 30000 {
		return fmt.Sprintf("consumables-30000/%s", i.Filename())
	}
	if i.ItemId >= 20000 {
		return fmt.Sprintf("armor-20000/%s/%s", i.Type, i.Filename())
	}
	if i.ItemId >= 10000 {
		return fmt.Sprintf("weapons-10000/%s", i.Filename())
	}
	return fmt.Sprintf("other-0/%s", i.Filename())
}

func (i ItemSpec) GetScript() string {

	scriptPath := i.GetScriptPath()

	// Load the script into a string
	if _, err := os.Stat(scriptPath); err == nil {
		if bytes, err := os.ReadFile(scriptPath); err == nil {
			return string(bytes)
		}
	}

	return ``
}

func (i *ItemSpec) GetScriptPath() string {
	// Load any script for the room
	return strings.Replace(string(configs.GetConfig().FolderItemData)+`/`+i.Filepath(), `.yaml`, `.js`, 1)
}

func GetItemSpec(itemId int) *ItemSpec {
	if itemId > 0 {
		spec, ok := items[itemId]
		if ok {
			return spec
		}
	}
	return nil
}

// file self loads due to init()
func LoadDataFiles() {

	start := time.Now()

	var err error
	items, err = fileloader.LoadAllFlatFiles[int, *ItemSpec](string(configs.GetConfig().FolderItemData))
	if err != nil {
		panic(err)
	}

	attackMessages, err = fileloader.LoadAllFlatFiles[ItemSubType, *WeaponAttackMessageGroup](string(configs.GetConfig().FolderAttackMessageData))
	if err != nil {
		panic(err)
	}

	slog.Info("itemspec.LoadDataFiles()", "itemLoadedCount", len(items), "attackMessageCount", len(attackMessages), "Time Taken", time.Since(start))

}
