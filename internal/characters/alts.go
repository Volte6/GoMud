package characters

import (
	"log/slog"
	"os"
	"strings"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/util"
	"gopkg.in/yaml.v2"
)

func AltsExists(username string) bool {
	_, err := os.Stat(util.FilePath(string(configs.GetConfig().FolderDataFiles), `/users/`, strings.ToLower(username)+`-alts.yaml`))

	return !os.IsNotExist(err)
}

func LoadAlts(username string) []Character {

	if !AltsExists(username) {
		return nil
	}

	altsFilePath := util.FilePath(string(configs.GetConfig().FolderDataFiles), `/users/`, strings.ToLower(username)+`-alts.yaml`)

	altsFileBytes, err := os.ReadFile(altsFilePath)
	if err != nil {
		slog.Error("LoadAlts", "error", err.Error())
		return nil
	}

	altsRecords := []Character{}

	if err := yaml.Unmarshal(altsFileBytes, &altsRecords); err != nil {
		slog.Error("LoadAlts", "error", err.Error())
	}

	return altsRecords

}

func SaveAlts(username string, alts []Character) bool {

	fileWritten := false
	tmpSaved := false
	tmpCopied := false
	completed := false

	defer func() {
		slog.Info("SaveAlts()", "username", username, "wrote-file", fileWritten, "tmp-file", tmpSaved, "tmp-copied", tmpCopied, "completed", completed)
	}()

	data, err := yaml.Marshal(&alts)
	if err != nil {
		slog.Error("SaveAlts", "error", err.Error())
		return false
	}

	carefulSave := configs.GetConfig().CarefulSaveFiles

	path := util.FilePath(string(configs.GetConfig().FolderDataFiles), `/users/`, strings.ToLower(username)+`-alts.yaml`)

	saveFilePath := path
	if carefulSave { // careful save first saves a {filename}.new file
		saveFilePath += `.new`
	}

	err = os.WriteFile(saveFilePath, data, 0777)
	if err != nil {
		slog.Error("SaveAlts", "error", err.Error())
		return false
	}
	fileWritten = true
	if carefulSave {
		tmpSaved = true
	}

	if carefulSave {
		//
		// Once the file is written, rename it to remove the .new suffix and overwrite the old file
		//
		if err := os.Rename(saveFilePath, path); err != nil {
			slog.Error("SaveAlts", "error", err.Error())
			return false
		}
		tmpCopied = true
	}

	completed = true

	return true

}
