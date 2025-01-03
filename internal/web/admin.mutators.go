package web

import (
	"log/slog"
	"net/http"
	"sort"
	"text/template"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mutators"
)

func mutatorsIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String()+"/admin/_header.html", configs.GetConfig().FolderHtmlFiles.String()+"/admin/mutators/index.html", configs.GetConfig().FolderHtmlFiles.String()+"/admin/_footer.html")
	if err != nil {
		slog.Error("HTML Template", "error", err)
	}

	mutSpecs := mutators.GetAllMutatorSpecs()

	sort.SliceStable(mutSpecs, func(i, j int) bool {
		return mutSpecs[i].MutatorId < mutSpecs[j].MutatorId
	})

	mutatorIndexData := struct {
		Mutators []mutators.MutatorSpec
	}{
		mutSpecs,
	}

	if err := tmpl.Execute(w, mutatorIndexData); err != nil {
		slog.Error("HTML Execute", "error", err)
	}

}

func mutatorData(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("mutator.data.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String() + "/admin/mutators/mutator.data.html")
	if err != nil {
		slog.Error("HTML Template", "error", err)
	}

	urlVals := r.URL.Query()

	mutatorId := urlVals.Get(`mutatorid`)

	mutSpec := mutators.GetMutatorSpec(mutatorId)
	if mutSpec == nil {
		mutSpec = &mutators.MutatorSpec{}
	}

	tplData := map[string]any{}
	tplData[`mutatorSpec`] = *mutSpec

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

	colorPatterns := colorpatterns.GetColorPatternNames()
	sort.SliceStable(colorPatterns, func(i, j int) bool {
		return colorPatterns[i] < colorPatterns[j]
	})
	tplData[`colorPatterns`] = colorPatterns

	if err := tmpl.Execute(w, tplData); err != nil {
		slog.Error("HTML Execute", "error", err)
	}

}
