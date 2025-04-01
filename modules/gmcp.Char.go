package modules

import (
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/plugins"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
)

// ////////////////////////////////////////////////////////////////////
// NOTE: The init function in Go is a special function that is
// automatically executed before the main function within a package.
// It is used to initialize variables, set up configurations, or
// perform any other setup tasks that need to be done before the
// program starts running.
// ////////////////////////////////////////////////////////////////////
func init() {

	//
	// We can use all functions only, but this demonstrates
	// how to use a struct
	//
	g := GMCPCharModule{
		plug: plugins.New(`gmcp.Char`, `1.0`),
	}
	g.plug.Requires(`gmcp`, `1.0`)

	// connectionId to map[string]int
	g.cache, _ = lru.New[uint64, map[string]int](256)

	events.RegisterListener(events.EquipmentChange{}, g.equipmentChangeHandler)
	events.RegisterListener(events.ItemOwnership{}, g.ownershipChangeHandler)

	events.RegisterListener(events.PlayerSpawn{}, g.playSpawnHandler)
	events.RegisterListener(events.CharacterVitalsChanged{}, g.vitalsChangedHandler)
	events.RegisterListener(events.LevelUp{}, g.levelUpHandler)
	events.RegisterListener(events.CharacterTrained{}, g.charTrainedHandler)
	events.RegisterListener(GMCPUpdate{}, g.buildAndSendGMCPPayload)
	events.RegisterListener(events.GainExperience{}, g.xpGainHandler)

	events.RegisterListener(GMCPModules{}, func(e events.Event) events.ListenerReturn {
		if evt, ok := e.(GMCPModules); ok {
			g.cache.Add(evt.ConnectionId, evt.Modules)
		}
		return events.Continue
	})

}

type GMCPCharModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug  *plugins.Plugin
	cache *lru.Cache[uint64, map[string]int]
}

// Tell the system a wish to send specific GMCP Update data
type GMCPUpdate struct {
	UserId     int
	Identifier string
}

func (g GMCPUpdate) Type() string { return `GMCPUpdate` }

func (g *GMCPCharModule) vitalsChangedHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.CharacterVitalsChanged)
	if !typeOk {
		return events.Continue // Return false to stop halt the event chain for this event
	}

	if evt.UserId == 0 {
		return events.Continue
	}

	// Changing equipment might affect stats, inventory, maxhp/maxmp etc
	events.AddToQueue(GMCPUpdate{
		UserId:     evt.UserId,
		Identifier: `Char.Vitals`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth
	})

	return events.Continue
}

func (g *GMCPCharModule) xpGainHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.GainExperience)
	if !typeOk {
		return events.Continue // Return false to stop halt the event chain for this event
	}

	if evt.UserId == 0 {
		return events.Continue
	}

	// Changing equipment might affect stats, inventory, maxhp/maxmp etc
	events.AddToQueue(GMCPUpdate{
		UserId:     evt.UserId,
		Identifier: `Char.Worth`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth
	})

	return events.Continue
}

func (g *GMCPCharModule) ownershipChangeHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.ItemOwnership)
	if !typeOk {
		return events.Continue // Return false to stop halt the event chain for this event
	}

	events.AddToQueue(GMCPUpdate{
		UserId:     evt.UserId,
		Identifier: `Char.Inventory`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth
	})

	return events.Continue
}

func (g *GMCPCharModule) equipmentChangeHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.EquipmentChange)
	if !typeOk {
		return events.Continue // Return false to stop halt the event chain for this event
	}

	if evt.UserId == 0 {
		return events.Continue
	}

	statsToChange := `Char`

	// If only gold or bank changed
	if len(evt.ItemsRemoved) == 0 && len(evt.ItemsWorn) == 0 {
		if evt.BankChange != 0 || evt.GoldChange != 0 {
			statsToChange = `Char.Worth`
		}
	}

	// Changing equipment might affect stats, inventory, maxhp/maxmp etc
	events.AddToQueue(GMCPUpdate{
		UserId:     evt.UserId,
		Identifier: statsToChange, // char, char.info, char.inventory, char.stats, char.vitals, char.worth
	})

	return events.Continue
}

func (g *GMCPCharModule) charTrainedHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.CharacterTrained)
	if !typeOk {
		return events.Continue // Return false to stop halt the event chain for this event
	}

	if evt.UserId == 0 {
		return events.Continue
	}

	// Changing equipment might affect stats, inventory, maxhp/maxmp etc
	events.AddToQueue(GMCPUpdate{
		UserId:     evt.UserId,
		Identifier: `Char`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth
	})

	return events.Continue
}
func (g *GMCPCharModule) levelUpHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.LevelUp)
	if !typeOk {
		return events.Continue // Return false to stop halt the event chain for this event
	}

	if evt.UserId == 0 {
		return events.Continue
	}

	// Changing equipment might affect stats, inventory, maxhp/maxmp etc
	events.AddToQueue(GMCPUpdate{
		UserId:     evt.UserId,
		Identifier: `Char`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth
	})

	return events.Continue
}

func (g *GMCPCharModule) playSpawnHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.PlayerSpawn)
	if !typeOk {
		return events.Continue // Return false to stop halt the event chain for this event
	}

	if evt.UserId == 0 {
		return events.Continue
	}

	// Send full update
	events.AddToQueue(GMCPUpdate{
		UserId:     evt.UserId,
		Identifier: `Char`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth
	})

	return events.Continue
}

// Checks whether their level is too high for a guide
func (g *GMCPCharModule) buildAndSendGMCPPayload(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(GMCPUpdate)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "GMCPUpdate", "Actual Type", e.Type())
		return events.Cancel
	}

	if evt.UserId < 1 {
		return events.Continue
	}

	// Make sure they have this gmcp module enabled.
	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return events.Continue
	}

	if len(evt.Identifier) >= 4 {

		identifierParts := strings.Split(strings.ToLower(evt.Identifier), `.`)
		for i := 0; i < len(identifierParts); i++ {
			identifierParts[i] = strings.Title(identifierParts[i])
		}

		rootModule := identifierParts[0]
		requestedId := strings.Join(identifierParts, `.`)

		if !g.supportsModule(user.ConnectionId(), rootModule) {
			return events.Continue
		}

		payload, moduleName := g.GetCharNode(user, requestedId)

		events.AddToQueue(GMCPOut{
			UserId:  evt.UserId,
			Module:  moduleName,
			Payload: payload,
		})

	}

	return events.Continue
}

func (g *GMCPCharModule) GetCharNode(user *users.UserRecord, gmcpModule string) (data any, moduleName string) {

	all := gmcpModule == `Char`

	payload := GMCPCharModule_Payload{}

	if all || g.wantsGMCPPayload(`Char.Info`, gmcpModule) {
		payload.Info = &GMCPCharModule_Payload_Info{
			Account:   user.Username,
			Name:      user.Character.Name,
			Class:     skills.GetProfession(user.Character.GetAllSkillRanks()),
			Race:      user.Character.Race(),
			Alignment: user.Character.AlignmentName(),
			Level:     user.Character.Level,
		}

		if !all {
			return payload.Info, `Char.Info`
		}
	}

	if all || g.wantsGMCPPayload(`Char.Inventory`, gmcpModule) {

		payload.Inventory = &GMCPCharModule_Payload_Inventory{

			Backpack: &GMCPCharModule_Payload_Inventory_Backpack{
				Count: len(user.Character.Items),
				Items: []string{},
				Max:   user.Character.CarryCapacity(),
			},

			Worn: &GMCPCharModule_Payload_Inventory_Worn{
				Weapon:  user.Character.Equipment.Weapon.Name(),
				Offhand: user.Character.Equipment.Offhand.Name(),
				Head:    user.Character.Equipment.Head.Name(),
				Neck:    user.Character.Equipment.Neck.Name(),
				Body:    user.Character.Equipment.Body.Name(),
				Belt:    user.Character.Equipment.Belt.Name(),
				Gloves:  user.Character.Equipment.Gloves.Name(),
				Ring:    user.Character.Equipment.Ring.Name(),
				Legs:    user.Character.Equipment.Legs.Name(),
				Feet:    user.Character.Equipment.Feet.Name(),
			},
		}

		// Fill the items list
		for _, itm := range user.Character.Items {
			payload.Inventory.Backpack.Items = append(payload.Inventory.Backpack.Items, itm.Name())
		}

		if !all {
			return payload.Inventory, `Char.Inventory`
		}
	}

	if all || g.wantsGMCPPayload(`Char.Stats`, gmcpModule) {

		payload.Stats = &GMCPCharModule_Payload_Stats{
			Strength:   user.Character.Stats.Strength.ValueAdj,
			Speed:      user.Character.Stats.Speed.ValueAdj,
			Smarts:     user.Character.Stats.Smarts.ValueAdj,
			Vitality:   user.Character.Stats.Vitality.ValueAdj,
			Mysticism:  user.Character.Stats.Mysticism.ValueAdj,
			Perception: user.Character.Stats.Perception.ValueAdj,
		}

		if !all {
			return payload.Stats, `Char.Stats`
		}
	}

	if all || g.wantsGMCPPayload(`Char.Vitals`, gmcpModule) {

		payload.Vitals = &GMCPCharModule_Payload_Vitals{
			Hp:    user.Character.Health,
			HpMax: user.Character.HealthMax.Value,
			Sp:    user.Character.Mana,
			SpMax: user.Character.ManaMax.Value,
		}

		if !all {
			return payload.Vitals, `Char.Vitals`
		}
	}

	if all || g.wantsGMCPPayload(`Char.Worth`, gmcpModule) {

		payload.Worth = &GMCPCharModule_Payload_Worth{
			Gold:           user.Character.Gold,
			Bank:           user.Character.Bank,
			SkillPoints:    user.Character.StatPoints,
			TrainingPoints: user.Character.TrainingPoints,
			TNL:            user.Character.XPTL(user.Character.Level),
			XP:             user.Character.Experience,
		}

		if !all {
			return payload.Worth, `Char.Worth`
		}
	}

	// If we reached this point and Char wasn't requested, we have a problem.
	if !all {
		mudlog.Error(`gmcp.Char`, `error`, `Bad module requested`, `module`, gmcpModule)
	}

	return payload, `Char`
}

// wantsGMCPPayload(`Char.Info`, `Char`)
func (g *GMCPCharModule) wantsGMCPPayload(searchName string, identifier string) bool {

	if searchName == identifier {
		return true
	}

	if len(searchName) < len(identifier) {
		return false
	}

	if searchName[0:len(identifier)] == identifier {
		return true
	}

	return false
}

func (g *GMCPCharModule) supportsModule(connectionId uint64, moduleName string) bool {
	if supportedModules, ok := g.cache.Get(connectionId); ok {
		if _, ok := supportedModules[moduleName]; ok {
			return true
		}
	}
	return false
}

type GMCPCharModule_Payload struct {
	Info      *GMCPCharModule_Payload_Info      `json:"Info,omitempty"`
	Inventory *GMCPCharModule_Payload_Inventory `json:"Inventory,omitempty"`
	Stats     *GMCPCharModule_Payload_Stats     `json:"Stats,omitempty"`
	Vitals    *GMCPCharModule_Payload_Vitals    `json:"Vitals,omitempty"`
	Worth     *GMCPCharModule_Payload_Worth     `json:"Worth,omitempty"`
}

// /////////////////
// Char.Info
// /////////////////
type GMCPCharModule_Payload_Info struct {
	Account   string `json:"account,omitempty"`
	Name      string `json:"name,omitempty"`
	Class     string `json:"class,omitempty"`
	Race      string `json:"race,omitempty"`
	Alignment string `json:"alignment,omitempty"`
	Level     int    `json:"level,omitempty"`
}

// /////////////////
// Char.Inventory
// /////////////////
type GMCPCharModule_Payload_Inventory struct {
	Backpack *GMCPCharModule_Payload_Inventory_Backpack `json:"Backpack,omitempty"`
	Worn     *GMCPCharModule_Payload_Inventory_Worn     `json:"Worn"`
}

type GMCPCharModule_Payload_Inventory_Backpack struct {
	Count int      `json:"count,omitempty"`
	Items []string `json:"items,omitempty"`
	Max   int      `json:"max,omitempty"`
}

type GMCPCharModule_Payload_Inventory_Worn struct {
	Weapon  string `json:"weapon,omitempty"`
	Offhand string `json:"offhand,omitempty"`
	Head    string `json:"head,omitempty"`
	Neck    string `json:"neck,omitempty"`
	Body    string `json:"body,omitempty"`
	Belt    string `json:"belt,omitempty"`
	Gloves  string `json:"gloves,omitempty"`
	Ring    string `json:"ring,omitempty"`
	Legs    string `json:"legs,omitempty"`
	Feet    string `json:"feet,omitempty"`
}

// /////////////////
// Char.Stats
// /////////////////
type GMCPCharModule_Payload_Stats struct {
	Strength   int `json:"strength,omitempty"`
	Speed      int `json:"speed,omitempty"`
	Smarts     int `json:"smarts,omitempty"`
	Vitality   int `json:"vitality,omitempty"`
	Mysticism  int `json:"mysticism,omitempty"`
	Perception int `json:"perception,omitempty"`
}

// /////////////////
// Char.Vitals
// /////////////////
type GMCPCharModule_Payload_Vitals struct {
	Hp    int `json:"hp,omitempty"`
	HpMax int `json:"hp_max,omitempty"`
	Sp    int `json:"sp,omitempty"`
	SpMax int `json:"sp_max,omitempty"`
}

// /////////////////
// Char.Worth
// /////////////////
type GMCPCharModule_Payload_Worth struct {
	Gold           int `json:"gold_carry,omitempty"`
	Bank           int `json:"gold_bank,omitempty"`
	SkillPoints    int `json:"skillpoints,omitempty"`
	TrainingPoints int `json:"trainingpoints,omitempty"`
	TNL            int `json:"tnl,omitempty"`
	XP             int `json:"xp,omitempty"`
}
