package modules

import (
	"strings"

	"github.com/volte6/gomud/internal/connections"
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

	events.RegisterListener(events.EquipmentChange{}, g.equipmentChangeHandler)
	events.RegisterListener(events.PlayerSpawn{}, g.playSpawnHandler)
	events.RegisterListener(events.CharacterVitalsChanged{}, g.vitalsChangedHandler)
	events.RegisterListener(events.LevelUp{}, g.levelUpHandler)
	events.RegisterListener(events.CharacterTrained{}, g.charTrainedHandler)
	events.RegisterListener(GMCPUpdate{}, g.buildAndSendGMCPPayload)

}

type GMCPCharModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug *plugins.Plugin
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
		Identifier: `char.vitals`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth *Can comma seaparate for multiple*
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

	statsToChange := ``

	if len(evt.ItemsRemoved) > 0 || len(evt.ItemsWorn) > 0 {
		statsToChange += `char.inventory, char.stats, char.vitals`
	}

	if evt.BankChange != 0 || evt.GoldChange != 0 {
		statsToChange += `, char.worth`
	}

	// Changing equipment might affect stats, inventory, maxhp/maxmp etc
	events.AddToQueue(GMCPUpdate{
		UserId:     evt.UserId,
		Identifier: statsToChange, // char, char.info, char.inventory, char.stats, char.vitals, char.worth *Can comma seaparate for multiple*
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
		Identifier: `char`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth *Can comma seaparate for multiple*
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
		Identifier: `char`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth *Can comma seaparate for multiple*
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
		Identifier: `char`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth *Can comma seaparate for multiple*
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
	if user == nil || !connections.GetClientSettings(user.ConnectionId()).GmcpEnabled(`Char`) {
		return events.Continue
	}

	if len(evt.Identifier) >= 4 {

		requestedId := strings.ToLower(evt.Identifier)

		payload := GMCPCharModule_Payload{}

		for _, requestIdPart := range strings.Split(requestedId, `,`) {

			requestIdPart = strings.TrimSpace(requestIdPart)
			if len(requestIdPart) == 0 {
				continue
			}

			// Info
			if g.wantsGMCPPayload(`char.info`, requestIdPart) {
				payload.Info.Account = user.Username
				payload.Info.Name = user.Character.Name
				payload.Info.Class = skills.GetProfession(user.Character.GetAllSkillRanks())
				payload.Info.Race = user.Character.Race()
				payload.Info.Alignment = user.Character.AlignmentName()
				payload.Info.Level = user.Character.Level
			}

			// Inventory
			if g.wantsGMCPPayload(`char.inventory`, requestIdPart) {
				// // Backpack
				payload.Inventory.Backpack.Count = len(user.Character.Items)
				payload.Inventory.Backpack.Items = []string{}
				for _, itm := range user.Character.Items {
					payload.Inventory.Backpack.Items = append(payload.Inventory.Backpack.Items, itm.Name())
				}
				payload.Inventory.Backpack.Max = user.Character.CarryCapacity()
				// // Worn
				payload.Inventory.Worn.Weapon = user.Character.Equipment.Weapon.Name()
				payload.Inventory.Worn.Offhand = user.Character.Equipment.Offhand.Name()
				payload.Inventory.Worn.Head = user.Character.Equipment.Head.Name()
				payload.Inventory.Worn.Neck = user.Character.Equipment.Neck.Name()
				payload.Inventory.Worn.Body = user.Character.Equipment.Body.Name()
				payload.Inventory.Worn.Belt = user.Character.Equipment.Belt.Name()
				payload.Inventory.Worn.Gloves = user.Character.Equipment.Gloves.Name()
				payload.Inventory.Worn.Ring = user.Character.Equipment.Ring.Name()
				payload.Inventory.Worn.Legs = user.Character.Equipment.Legs.Name()
				payload.Inventory.Worn.Feet = user.Character.Equipment.Feet.Name()
			}

			if g.wantsGMCPPayload(`char.stats`, requestIdPart) {
				// Stats
				payload.Stats.Strength = user.Character.Stats.Strength.ValueAdj
				payload.Stats.Speed = user.Character.Stats.Speed.ValueAdj
				payload.Stats.Smarts = user.Character.Stats.Smarts.ValueAdj
				payload.Stats.Vitality = user.Character.Stats.Vitality.ValueAdj
				payload.Stats.Mysticism = user.Character.Stats.Mysticism.ValueAdj
				payload.Stats.Perception = user.Character.Stats.Perception.ValueAdj
			}

			if g.wantsGMCPPayload(`char.vitals`, requestIdPart) {
				// Vitals
				payload.Vitals.Hp = user.Character.Health
				payload.Vitals.HpMax = user.Character.HealthMax.Value
				payload.Vitals.Sp = user.Character.Mana
				payload.Vitals.SpMax = user.Character.ManaMax.Value
			}

			if g.wantsGMCPPayload(`char.worth`, requestIdPart) {
				// Worth
				payload.Worth.Gold = user.Character.Gold
				payload.Worth.Bank = user.Character.Bank
				payload.Worth.SkillPoints = user.Character.StatPoints
				payload.Worth.TrainingPoints = user.Character.TrainingPoints
				payload.Worth.TNL = user.Character.XPTL(user.Character.Level + 1)
				payload.Worth.XP = user.Character.Experience
			}
		}

		events.AddToQueue(GMCPOut{
			UserId:  evt.UserId,
			Module:  `Char`,
			Payload: payload,
		})

	}

	return events.Continue
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

type GMCPCharModule_Payload struct {
	//
	Info struct {
		Account   string `json:"account,omitempty"`
		Name      string `json:"name,omitempty"`
		Class     string `json:"class,omitempty"`
		Race      string `json:"race,omitempty"`
		Alignment string `json:"alignment,omitempty"`
		Level     int    `json:"level,omitempty"`
	} `json:"Info,omitempty"`
	//
	Inventory struct {
		//
		Backpack struct {
			Count int      `json:"count,omitempty"`
			Items []string `json:"items,omitempty"`
			Max   int      `json:"max,omitempty"`
		} `json:"Backpack,omitempty"`

		Worn struct {
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
		} `json:"Worn"`
		//
	} `json:"Inventory,omitempty"`
	//
	Stats struct {
		Strength   int `json:"strength,omitempty"`
		Speed      int `json:"speed,omitempty"`
		Smarts     int `json:"smarts,omitempty"`
		Vitality   int `json:"vitality,omitempty"`
		Mysticism  int `json:"mysticism,omitempty"`
		Perception int `json:"perception,omitempty"`
	} `json:"Stats,omitempty"`
	//
	Vitals struct {
		Hp    int `json:"hp,omitempty"`
		HpMax int `json:"hp_max,omitempty"`
		Sp    int `json:"sp,omitempty"`
		SpMax int `json:"sp_max,omitempty"`
	} `json:"Vitals,omitempty"`
	//
	Worth struct {
		Gold           int `json:"gold_carry,omitempty"`
		Bank           int `json:"gold_bank,omitempty"`
		SkillPoints    int `json:"skillpoints,omitempty"`
		TrainingPoints int `json:"trainingpoints,omitempty"`
		TNL            int `json:"tnl,omitempty"`
		XP             int `json:"xp,omitempty"`
	} `json:"Worth,omitempty"`
}
