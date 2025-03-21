package plugins

import (
	"embed"
	"io/fs"
)

// Implements fs.ReadFileFS
type PluginFiles struct {
	fileSystem embed.FS
	filePaths  map[string]string
}

func (p PluginFiles) ReadFile(name string) ([]byte, error) {
	for _, p := range registry {

		if embedPath, ok := p.files.filePaths[name]; ok {
			b, err := p.files.fileSystem.ReadFile(embedPath)
			if err == nil {
				return b, nil
			}
		}
	}

	return nil, fs.ErrNotExist
}

func (p PluginFiles) Open(name string) (fs.File, error) {

	for _, p := range registry {

		if embedPath, ok := p.files.filePaths[name]; ok {
			return p.files.fileSystem.Open(embedPath)

		}
	}

	return nil, fs.ErrNotExist

}

func (p PluginFiles) Stat(name string) (fs.FileInfo, error) {

	for _, p := range registry {

		if embedPath, ok := p.files.filePaths[name]; ok {
			return fs.Stat(p.files.fileSystem, embedPath)
		}
	}

	return nil, fs.ErrNotExist

}
