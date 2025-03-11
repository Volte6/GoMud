package spells

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/fileloader"

	"github.com/volte6/gomud/internal/util"
)

func CreateNewSpellFile(newSpellInfo SpellData) (string, error) {

	if sp := GetSpell(newSpellInfo.SpellId); sp != nil {
		return ``, errors.New(`Spell already exists.`)
	}

	if err := newSpellInfo.Validate(); err != nil {
		return ``, err
	}

	saveModes := []fileloader.SaveOption{}

	if configs.GetFilePathsConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}

	if err := fileloader.SaveFlatFile[*SpellData](string(configs.GetFilePathsConfig().FolderDataFiles)+`/spells`, &newSpellInfo, saveModes...); err != nil {
		return ``, err
	}

	// Save to in-memory cache
	allSpells[newSpellInfo.Id()] = &newSpellInfo

	newScriptPath := newSpellInfo.GetScriptPath()
	os.MkdirAll(filepath.Dir(newScriptPath), os.ModePerm)

	fileloader.CopyFileContents(util.FilePath(`_datafiles/sample-scripts/spells/`+string(newSpellInfo.Type)+`.js`),
		newScriptPath)

	return newSpellInfo.SpellId, nil
}
