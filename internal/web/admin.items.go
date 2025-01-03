package web

import (
	"html"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"text/template"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/items"
)

func itemsIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String()+"/admin/_header.html", configs.GetConfig().FolderHtmlFiles.String()+"/admin/items/index.html", configs.GetConfig().FolderHtmlFiles.String()+"/admin/_footer.html")
	if err != nil {
		slog.Error("HTML Template", "error", err)
	}

	qsp := r.URL.Query()

	filterType := qsp.Get(`filter-type`)

	itemSpecs := []items.ItemSpec{}

	itemTypes := items.ItemTypes()
	itemTypes = append(itemTypes, items.ItemSubtypes()...)

	typeCounter := map[string]int{}

	for _, itemSpec := range items.GetAllItemSpecs() {

		typeCounter[itemSpec.Type.String()] += 1
		typeCounter[itemSpec.Subtype.String()] += 1

		if filterType != `*` && filterType != itemSpec.Type.String() && filterType != itemSpec.Subtype.String() {
			continue
		}

		itemSpecs = append(itemSpecs, itemSpec)
	}

	for i, typeInfo := range itemTypes {
		itemTypes[i].Count = typeCounter[typeInfo.Type]
	}

	sort.SliceStable(itemSpecs, func(i, j int) bool {
		return itemSpecs[i].ItemId < itemSpecs[j].ItemId
	})

	sort.SliceStable(itemTypes, func(i, j int) bool {
		return itemTypes[i].Count > itemTypes[j].Count
	})

	itemIndexData := struct {
		ItemSpecs  []items.ItemSpec
		ItemTypes  []items.ItemTypeInfo
		FilterType string
	}{
		itemSpecs,
		itemTypes,
		filterType,
	}

	if err := tmpl.Execute(w, itemIndexData); err != nil {
		slog.Error("HTML Execute", "error", err)
	}

}

func itemData(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("item.data.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String() + "/admin/items/item.data.html")
	if err != nil {
		slog.Error("HTML Template", "error", err)
	}

	urlVals := r.URL.Query()

	itemInt, _ := strconv.Atoi(urlVals.Get(`itemid`))

	itemSpec := items.GetItemSpec(itemInt)
	if itemSpec == nil {
		itemSpec = &items.ItemSpec{}
	}

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

	if err := tmpl.Execute(w, tplData); err != nil {
		slog.Error("HTML Execute", "error", err)
	}

}
