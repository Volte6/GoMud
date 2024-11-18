package web

import (
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"text/template"

	"github.com/volte6/gomud/buffs"
	"github.com/volte6/gomud/characters"
	"github.com/volte6/gomud/mobs"
	"github.com/volte6/gomud/races"
	"github.com/volte6/gomud/rooms"
)

func mobsIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("_datafiles/html/admin/_header.html", "_datafiles/html/admin/mobs/index.html", "_datafiles/html/admin/_footer.html")
	if err != nil {
		slog.Error("HTML Template", "error", err)
	}

	allMobs := mobs.GetAllMobInfo()
	sort.SliceStable(allMobs, func(i, j int) bool {
		return allMobs[i].MobId < allMobs[j].MobId
	})

	mobIndexData := struct {
		Mobs []mobs.Mob
	}{
		allMobs,
	}

	if err := tmpl.Execute(w, mobIndexData); err != nil {
		slog.Error("HTML Execute", "error", err)
	}

}

func mobData(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("mob.data.html").Funcs(funcMap).ParseFiles("_datafiles/html/admin/mobs/mob.data.html")
	if err != nil {
		slog.Error("HTML Template", "error", err)
	}

	urlVals := r.URL.Query()

	mobIdInt, _ := strconv.Atoi(urlVals.Get(`mobid`))

	mobInfo := mobs.GetMobSpec(mobs.MobId(mobIdInt))
	if mobInfo == nil {
		mobInfo = &mobs.Mob{}
	}

	mobGroupSet := map[string]struct{}{}
	allMobGroups := []string{}
	for _, m := range mobs.GetAllMobInfo() {

		for _, groupName := range m.Groups {
			if _, ok := mobGroupSet[groupName]; !ok {
				allMobGroups = append(allMobGroups, groupName)
				mobGroupSet[groupName] = struct{}{}
			}
		}

	}

	allRaces := races.GetRaces()
	sort.SliceStable(allRaces, func(i, j int) bool {
		return allRaces[i].RaceId < allRaces[j].RaceId
	})

	allZoneNames := rooms.GetAllZoneNames()
	sort.SliceStable(allZoneNames, func(i, j int) bool {
		return allZoneNames[i] < allZoneNames[j]
	})

	activityLevels := []int{}
	for i := 0; i < 11; i++ {
		activityLevels = append(activityLevels, i)
	}

	dropChances := []int{}
	for i := 0; i < 101; i++ {
		dropChances = append(dropChances, i)
	}

	buffSpecs := []buffs.BuffSpec{}
	for _, buffId := range buffs.GetAllBuffIds() {
		if b := buffs.GetBuffSpec(buffId); b != nil {
			if b.Name == `empty` {
				continue
			}
			buffSpecs = append(buffSpecs, *b)
		}
	}
	sort.SliceStable(buffSpecs, func(i, j int) bool {
		return buffSpecs[i].BuffId < buffSpecs[j].BuffId
	})

	tplData := map[string]any{}

	tplData[`mobInfo`] = *mobInfo

	shopData := map[string]characters.Shop{
		`Items`:       {},
		`Buffs`:       {},
		`Mercenaries`: {},
		`Pets`:        {},
	}

	for _, shopItm := range mobInfo.Character.Shop {

		if shopItm.ItemId > 0 {
			shopData[`Items`] = append(shopData[`Items`], shopItm)
			continue
		}

		if shopItm.BuffId > 0 {
			shopData[`Buffs`] = append(shopData[`Buffs`], shopItm)
			continue
		}

		if shopItm.MobId > 0 {
			shopData[`Mercenaries`] = append(shopData[`Mercenaries`], shopItm)
			continue
		}

		if shopItm.PetType != `` {
			shopData[`Pets`] = append(shopData[`Pets`], shopItm)
			continue
		}
	}
	tplData[`mobShop`] = shopData

	tplData[`characterInfo`] = &mobInfo.Character
	tplData[`allZoneNames`] = allZoneNames
	tplData[`allRaces`] = allRaces
	tplData[`activityLevels`] = activityLevels
	tplData[`dropChances`] = dropChances
	tplData[`allMobGroups`] = allMobGroups
	tplData[`buffSpecs`] = buffSpecs

	if err := tmpl.Execute(w, tplData); err != nil {
		slog.Error("HTML Execute", "error", err)
	}

}
