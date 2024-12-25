package mobs

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/fileloader"
	"github.com/volte6/gomud/internal/util"
)

var (
	SampleScripts = map[string]string{
		`item and gold`: `item-gold-quest.js`,
	}
)

const (
	ScriptTemplateQuest = `item-gold-quest.js`
)

func CreateNewMobFile(newMobInfo Mob, copyScript string) (MobId, error) {

	newMobInfo.MobId = getNextMobId()

	if newMobInfo.MobId == 0 {
		return 0, errors.New(`Could not find a new mob id to assign.`)
	}

	if copyScript == ScriptTemplateQuest {
		newMobInfo.QuestFlags = []string{`1000000-start`}
	}

	if err := newMobInfo.Validate(); err != nil {
		return 0, err
	}

	if newMobInfo.MobId == 0 {
		return 0, errors.New(`Could not find a new mob id to assign.`)
	}

	saveModes := []fileloader.SaveOption{}

	if configs.GetConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}

	if err := fileloader.SaveFlatFile[*Mob](mobDataFilesFolderPath, &newMobInfo, saveModes...); err != nil {
		return 0, err
	}

	// Save to in-memory cache
	allMobNames = append(allMobNames, newMobInfo.Character.Name)
	mobNameCache[newMobInfo.MobId] = newMobInfo.Character.Name
	mobs[newMobInfo.Id()] = &newMobInfo

	if copyScript != `` {

		newScriptPath := newMobInfo.GetScriptPath()
		os.MkdirAll(filepath.Dir(newScriptPath), os.ModePerm)

		fileloader.CopyFileContents(
			util.FilePath(string(configs.GetConfig().FolderSampleScripts)+`/mobs/`+copyScript),
			newMobInfo.GetScriptPath(),
		)
	}

	return newMobInfo.MobId, nil
}

func getNextMobId() MobId {

	lowestFreeId := MobId(0)
	for _, mInfo := range mobs {

		if mInfo.MobId >= lowestFreeId {
			lowestFreeId = mInfo.MobId + 1
		}
	}

	return lowestFreeId
}
