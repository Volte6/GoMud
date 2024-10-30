package web

import (
	"html"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"text/template"

	"github.com/volte6/gomud/buffs"
	"github.com/volte6/gomud/items"
	"github.com/volte6/gomud/util"
)

func itemsIndex(w http.ResponseWriter, r *http.Request) {

	allItemSpecs := items.GetAllItemSpecs()

	sort.SliceStable(allItemSpecs, func(i, j int) bool {
		return allItemSpecs[i].ItemId < allItemSpecs[j].ItemId
	})

	tmpl, err := template.New("items.html").Funcs(funcMap).ParseFiles("web/html/admin/items.html")
	if err != nil {
		slog.Error("HTML ERROR 1", "error", err)
	}

	if err := tmpl.Execute(w, allItemSpecs); err != nil {
		slog.Error("HTML ERROR 2", "error", err)
	}

}

func itemData(w http.ResponseWriter, r *http.Request) {

	urlVals := r.URL.Query()

	itemInt, _ := strconv.Atoi(urlVals.Get(`itemid`))

	util.LockGame()
	defer util.UnlockGame()

	if itemSpec := items.GetItemSpec(itemInt); itemSpec != nil {

		tplData := map[string]any{}
		tplData[`itemSpec`] = *itemSpec

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

		tplData[`itemTypes`] = items.ItemTypes()
		tplData[`itemSubtypes`] = items.ItemSubtypes()

		tplData[`script`] = html.EscapeString(itemSpec.GetScript())

		tmpl, err := template.New("itemdata.html").Funcs(funcMap).ParseFiles("web/html/admin/itemdata.html")
		if err != nil {
			slog.Error("HTML ERROR 1", "error", err)
		}

		if err := tmpl.Execute(w, tplData); err != nil {
			slog.Error("HTML ERROR 2", "error", err)
		}

		return

	}
	w.Write([]byte("Not found: " + urlVals.Get(`itemid`)))
}
