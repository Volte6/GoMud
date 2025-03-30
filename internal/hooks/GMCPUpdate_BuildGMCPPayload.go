package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
)

// This is temporary, remove after done testing.
func ForceGMCPCharUpdate(e events.Event) events.ListenerReturn {

	evt, _ := e.(events.NewRound)
	for _, userId := range users.GetOnlineUserIds() {

		ask := `Char`

		switch evt.RoundNumber % 6 {
		case 0:
			ask = `Char`
		case 1:
			ask = `Char.Info`
		case 2:
			ask = `Char.Inventory`
		case 3:
			ask = `Char.Stats`
		case 4:
			ask = `Char.Vitals`
		case 5:
			ask = `Char.Worth`
		}

		events.AddToQueue(events.GMCPUpdate{
			UserId:     userId,
			Identifier: ask,
		})

	}

	return events.Continue
}

// Checks whether their level is too high for a guide
func BuildGMCPPayload(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.GMCPUpdate)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "GMCPUpdate", "Actual Type", e.Type())
		return events.Cancel
	}

	if evt.UserId < 1 {
		return events.Continue
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return events.Continue
	}

	if len(evt.Identifier) >= 4 {

		payload := GMCPChar{}

		// Info
		if WantsGMCPPayload(`Char.Info`, evt.Identifier) {
			payload.Info.Account = user.Username
			payload.Info.Name = user.Character.Name
			payload.Info.Class = skills.GetProfession(user.Character.GetAllSkillRanks())
			payload.Info.Race = user.Character.Race()
			payload.Info.Alignment = user.Character.AlignmentName()
			payload.Info.Level = user.Character.Level
		}

		// Inventory
		if WantsGMCPPayload(`Char.Inventory`, evt.Identifier) {
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

		if WantsGMCPPayload(`Char.Stats`, evt.Identifier) {
			// Stats
			payload.Stats.Strength = user.Character.Stats.Strength.ValueAdj
			payload.Stats.Speed = user.Character.Stats.Speed.ValueAdj
			payload.Stats.Smarts = user.Character.Stats.Smarts.ValueAdj
			payload.Stats.Vitality = user.Character.Stats.Vitality.ValueAdj
			payload.Stats.Mysticism = user.Character.Stats.Mysticism.ValueAdj
			payload.Stats.Perception = user.Character.Stats.Perception.ValueAdj
		}

		if WantsGMCPPayload(`Char.Vitals`, evt.Identifier) {
			// Vitals
			payload.Vitals.Hp = user.Character.Health
			payload.Vitals.HpMax = user.Character.HealthMax.Value
			payload.Vitals.Sp = user.Character.Mana
			payload.Vitals.SpMax = user.Character.ManaMax.Value
		}

		if WantsGMCPPayload(`Char.Worth`, evt.Identifier) {
			// Worth
			payload.Worth.Gold = user.Character.Gold
			payload.Worth.Bank = user.Character.Bank
			payload.Worth.SkillPoints = user.Character.StatPoints
			payload.Worth.TrainingPoints = user.Character.TrainingPoints
			payload.Worth.TNL = user.Character.XPTL(user.Character.Level + 1)
			payload.Worth.XP = user.Character.Experience
		}

		events.AddToQueue(events.GMCPOut{
			UserId:  evt.UserId,
			Payload: payload,
		})

	}

	return events.Continue
}

// WantsGMCPPayload(`Char.Info`, `Char`)
func WantsGMCPPayload(searchName string, identifier string) bool {

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

type GMCPChar struct {
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
