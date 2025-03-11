package web

import (
	"net/http"
	"sort"
	"strconv"
	"text/template"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/races"
)

func racesIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles(configs.GetFilePathsConfig().FolderAdminHtml.String()+"/_header.html", configs.GetFilePathsConfig().FolderAdminHtml.String()+"/races/index.html", configs.GetFilePathsConfig().FolderAdminHtml.String()+"/_footer.html")
	if err != nil {
		mudlog.Error("HTML Template", "error", err)
	}

	allRaces := races.GetRaces()

	sort.SliceStable(allRaces, func(i, j int) bool {
		return allRaces[i].RaceId < allRaces[j].RaceId
	})

	raceIndexData := struct {
		Races []races.Race
	}{
		allRaces,
	}

	if err := tmpl.Execute(w, raceIndexData); err != nil {
		mudlog.Error("HTML Execute", "error", err)
	}

}

func raceData(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("race.data.html").Funcs(funcMap).ParseFiles(configs.GetFilePathsConfig().FolderAdminHtml.String() + "/races/race.data.html")
	if err != nil {
		mudlog.Error("HTML Template", "error", err)
	}

	urlVals := r.URL.Query()

	raceIdInt, _ := strconv.Atoi(urlVals.Get(`raceid`))

	raceInfo := races.GetRace(raceIdInt)
	if raceInfo == nil {
		raceInfo = &races.Race{}
	}

	tplData := map[string]any{}
	tplData[`raceInfo`] = *raceInfo

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
	tplData[`buffSpecs`] = buffSpecs

	tplData[`allSlotTypes`] = characters.GetAllSlotTypes()

	if err := tmpl.Execute(w, tplData); err != nil {
		mudlog.Error("HTML Execute", "error", err)
	}

}
