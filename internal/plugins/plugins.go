package plugins

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/volte6/gomud/internal/mobcommands"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/usercommands"
	"github.com/volte6/gomud/internal/util"
)

//
// This package defines the basic nature of GoMud plugins
// To add new plugins, they must be dropped in this folder and the server re-compiled.
//

// pluginRegistry holds all plugins, provides a `fs.ReadFileFS` interface
type pluginRegistry []*Plugin
type FileType int

var (
	registrationOpen = true
	registry         = pluginRegistry{}
	txtCleanRegex    = regexp.MustCompile(`[^a-zA-Z0-9\._]+`)
	writeFolderPath  = os.TempDir()
)

const (
	dataFilesFolder         = `datafiles` + string(filepath.Separator)
	dataOverlaysFilesFolder = `data-overlays` + string(filepath.Separator)
)

func (p pluginRegistry) ReadFile(name string) ([]byte, error) {
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

func (p pluginRegistry) Open(name string) (fs.File, error) {

	for _, p := range registry {

		if embedPath, ok := p.files.filePaths[name]; ok {
			return p.files.fileSystem.Open(embedPath)

		}
	}

	return nil, fs.ErrNotExist

}

func (p pluginRegistry) Stat(name string) (fs.FileInfo, error) {

	for _, p := range registry {

		if embedPath, ok := p.files.filePaths[name]; ok {
			return fs.Stat(p.files.fileSystem, embedPath)
		}
	}

	return nil, fs.ErrNotExist

}

//
// Plugin struct
//

type Plugin struct {
	name    string
	version string

	callbacks struct {
		userCommands map[string]usercommands.CommandAccess
		mobCommands  map[string]mobcommands.CommandAccess

		onLoad func()
		onSave func()
	}

	// helper for embedded files
	files struct {
		fileSystem embed.FS
		filePaths  map[string]string
	}
}

func New(name string, version string) *Plugin {

	if !registrationOpen {
		return nil
	}

	p := &Plugin{
		name:    name,
		version: version,
	}

	p.callbacks.userCommands = map[string]usercommands.CommandAccess{}
	p.callbacks.mobCommands = map[string]mobcommands.CommandAccess{}

	registry = append(registry, p)
	return p
}

// Registers a UserCommand and callback
func (p *Plugin) AddUserCommand(command string, handlerFunc usercommands.UserCommand, allowWhenDowned bool, isAdminOnly bool) {

	if p.callbacks.userCommands == nil {
		p.callbacks.userCommands = map[string]usercommands.CommandAccess{}
	}

	p.callbacks.userCommands[command] = usercommands.CommandAccess{
		Func:              handlerFunc,
		AllowedWhenDowned: allowWhenDowned,
		AdminOnly:         isAdminOnly,
	}
}

// Registers a MobCommand and callback
func (p *Plugin) AddMobCommand(command string, handlerFunc mobcommands.MobCommand, allowWhenDowned bool) {

	if p.callbacks.mobCommands == nil {
		p.callbacks.mobCommands = map[string]mobcommands.CommandAccess{}
	}

	p.callbacks.mobCommands[command] = mobcommands.CommandAccess{
		Func:              handlerFunc,
		AllowedWhenDowned: allowWhenDowned,
	}

}

// Adds an embedded file system to the plugin
func (p *Plugin) AttachFileSystem(f embed.FS) error {

	p.files.fileSystem = f

	p.files.filePaths = make(map[string]string)

	// Walk the directory tree rooted at "datafiles"
	err := fs.WalkDir(p.files.fileSystem, `.`, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // propagate the error
		}

		// If it's not a directory, add the path to the list
		if d.IsDir() {
			return nil
		}

		// Handle datafiles folder.
		dfPos := strings.Index(path, dataFilesFolder)
		if dfPos != -1 {
			// map the short path to long embedded path
			p.files.filePaths[path[dfPos+len(dataFilesFolder):]] = path
			return nil
		}

		// Handle data-overlays folder.
		// This is a special folder that overlays data onto other data
		dfPos = strings.Index(path, dataOverlaysFilesFolder)
		if dfPos != -1 {
			// map the short path to long embedded path
			// Put data-overlays/ prefix on for purposes of filefinding later.
			p.files.filePaths[dataOverlaysFilesFolder+path[dfPos+len(dataOverlaysFilesFolder):]] = path
			return nil
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("plug.AddFiles() for %s: %w", p.name, err)
	}

	return nil
}

func (p *Plugin) SetOnLoad(f func()) {
	p.callbacks.onLoad = f
}

func (p *Plugin) SetOnSave(f func()) {
	p.callbacks.onSave = f
}

func (p *Plugin) WriteBytes(identifier string, bytes []byte) error {

	// Fix up identifier
	fileName := strings.ToLower(txtCleanRegex.ReplaceAllString(identifier, "-")) + `.plugin.dat`

	// Fix up folderpath
	folderPath := util.FilePath(writeFolderPath, `/`, strings.ToLower(txtCleanRegex.ReplaceAllString(fmt.Sprintf(`%s-v%s`, p.name, p.version), "-")))

	// Create full path
	fullPath := util.FilePath(folderPath, `/`, fileName)

	if _, err := os.Stat(fullPath); err != nil {
		if err = os.MkdirAll(folderPath, 0777); err != nil {
			mudlog.Error(`plugin.WriteBytes`, `name`, p.name, `path`, folderPath, `error`, err)
			return err
		}
	}

	if err := os.WriteFile(fullPath, bytes, 0777); err != nil {
		mudlog.Error(`plugin.WriteBytes`, `name`, p.name, `path`, fullPath, `error`, err)
		return err
	}

	return nil
}

func (p *Plugin) ReadBytes(identifier string) ([]byte, error) {

	// Fix up identifier
	fileName := strings.ToLower(txtCleanRegex.ReplaceAllString(identifier, "-")) + `.plugin.dat`

	// Fix up folderpath
	folderPath := util.FilePath(writeFolderPath, `/`, strings.ToLower(txtCleanRegex.ReplaceAllString(fmt.Sprintf(`%s-v%s`, p.name, p.version), "-")))

	// Create full path
	fullPath := util.FilePath(folderPath, `/`, fileName)

	bytes, err := os.ReadFile(fullPath)
	if err != nil {
		mudlog.Warn(`plugin.ReadBytes`, `name`, p.name, `path`, fullPath, `error`, err)
	}

	return bytes, err
}

func Load(dataFilesPath string) {

	writeFolderPath = util.FilePath(dataFilesPath, `/`, `plugin-data`)

	registrationOpen = false

	pluginCt := 0
	for _, p := range registry {

		pluginCt++

		for cmd, info := range p.callbacks.userCommands {
			usercommands.RegisterCommand(cmd, info.Func, info.AllowedWhenDowned, info.AdminOnly)
		}

		for cmd, info := range p.callbacks.mobCommands {
			mobcommands.RegisterCommand(cmd, info.Func, info.AllowedWhenDowned)
		}

		if p.callbacks.onLoad != nil {
			p.callbacks.onLoad()
		}
	}

	mudlog.Info("plugins", "loadedCount", pluginCt)
}

func Save() {

	pluginCt := 0
	for _, p := range registry {

		if p.callbacks.onSave != nil {
			p.callbacks.onSave()
			pluginCt++
		}

	}
	mudlog.Info("plugins", "saveCount", pluginCt)
}

func GetRegistryFS() fs.ReadFileFS {
	return registry
}

func ReadFile(dfPath string) ([]byte, error) {
	return registry.ReadFile(dfPath)
}
