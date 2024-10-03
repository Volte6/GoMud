package fileloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"sync/atomic"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type FileType uint8
type SaveOption uint8

type LoadableSimple interface {
	Validate() error  // General validation (or none)
	Filepath() string // Relative file path to some base directory - can include subfolders
}

type Loadable[K comparable] interface {
	Id() K // Must be a unique identifier for the data
	LoadableSimple
}

const (
	// File types to load
	FileTypeYaml FileType = iota
	FileTypeJson

	// Save options
	SaveCareful SaveOption = iota // Save a backup and rename vs. just overwriting
)

func LoadFlatFile[T LoadableSimple](path string) (T, error) {

	var loaded T

	path = filepath.FromSlash(path)

	fileInfo, err := os.Stat(path)
	if err != nil {
		return loaded, errors.Wrap(err, `filepath: `+path)
	}

	if fileInfo.IsDir() {
		return loaded, errors.New(`filepath: ` + path + ` is a directory`)
	}

	fpathLower := strings.ToLower(path[len(path)-5:]) // Only need to compare the last 5 characters

	if fpathLower == `.yaml` {

		bytes, err := os.ReadFile(path)
		if err != nil {
			return loaded, errors.Wrap(err, `filepath: `+path)
		}

		err = yaml.Unmarshal(bytes, &loaded)
		if err != nil {
			return loaded, errors.Wrap(err, `filepath: `+path)
		}

	} else if fpathLower == `.json` {

		bytes, err := os.ReadFile(path)
		if err != nil {
			return loaded, errors.Wrap(err, `filepath: `+path)
		}

		err = json.Unmarshal(bytes, &loaded)
		if err != nil {
			return loaded, errors.Wrap(err, `filepath: `+path)
		}

	} else {
		// Skip the file altogether
		return loaded, errors.New(`invalid file type: ` + path)
	}

	// Make sure the Filepath it claims is correct in case we need to save it later
	if !strings.HasSuffix(path, filepath.FromSlash(loaded.Filepath())) {
		return loaded, errors.New(fmt.Sprintf(`filesystem path "%s" did not end in Filepath() "%s" for type %T`, path, loaded.Filepath(), loaded))
	}

	// validate the structure
	if err := loaded.Validate(); err != nil {
		return loaded, errors.Wrap(err, `filepath: `+path)
	}

	return loaded, nil
}

// LoadAllFlatFilesSimple doesn't require a unique Id() for each item
func LoadAllFlatFilesSimple[T LoadableSimple](basePath string, fileTypes ...FileType) ([]T, error) {

	loadedData := make([]T, 0, 128)

	includeYaml := true
	includeJson := true

	if len(fileTypes) > 0 {
		includeYaml = false
		includeJson = false

		for _, fType := range fileTypes {
			if fType == FileTypeYaml {
				includeYaml = true
			} else if fType == FileTypeJson {
				includeJson = true
			}
		}
	}

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// only lower the last 5 characters of the path string
		fpathLower := strings.ToLower(path[len(path)-5:])

		if !includeYaml && fpathLower == `.yaml` {
			return errors.New(`invalid file type (yaml): ` + path)
		}

		if !includeJson && fpathLower == `.json` {
			return errors.New(`invalid file type (json): ` + path)
		}

		loaded, err := LoadFlatFile[T](path)

		if err != nil {
			return err
		}

		loadedData = append(loadedData, loaded)

		return nil
	})

	return loadedData, err
}

// Will check the ID() of each item to make sure it's unique
func LoadAllFlatFiles[K comparable, T Loadable[K]](basePath string, fileTypes ...FileType) (map[K]T, error) {

	basePath = filepath.FromSlash(basePath)

	loadedData := make(map[K]T)

	includeYaml := true
	includeJson := true

	if len(fileTypes) > 0 {
		includeYaml = false
		includeJson = false

		for _, fType := range fileTypes {
			if fType == FileTypeYaml {
				includeYaml = true
			} else if fType == FileTypeJson {
				includeJson = true
			}
		}
	}

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		var loaded T

		fpathLower := path[len(path)-5:] // Only need to compare the last 5 characters
		if includeYaml && fpathLower == `.yaml` {

			bytes, err := os.ReadFile(path)
			if err != nil {
				return errors.Wrap(err, `filepath: `+path)
			}

			err = yaml.Unmarshal(bytes, &loaded)
			if err != nil {
				return errors.Wrap(err, `filepath: `+path)
			}

		} else if includeJson && fpathLower == `.json` {

			bytes, err := os.ReadFile(path)
			if err != nil {
				return errors.Wrap(err, `filepath: `+path)
			}

			err = json.Unmarshal(bytes, &loaded)
			if err != nil {
				return errors.Wrap(err, `filepath: `+path)
			}

		} else {
			// Skip the file altogether
			return nil
		}

		if !strings.HasSuffix(path, filepath.FromSlash(loaded.Filepath())) {
			return errors.New(fmt.Sprintf(`filesystem path "%s" did not end in Filepath() "%s" for type %T`, path, loaded.Filepath(), loaded))
		}

		if err := loaded.Validate(); err != nil {
			return errors.Wrap(err, `filepath: `+path)
		}

		if _, ok := loadedData[loaded.Id()]; ok {
			return errors.New(fmt.Sprintf(`duplicate id %v for type %T`, loaded.Id(), loaded))
		}

		loadedData[loaded.Id()] = loaded

		return nil
	})

	return loadedData, err
}

// Returns the number of files saved and error
func SaveFlatFile[T LoadableSimple](basePath string, dataUnit T, saveOptions ...SaveOption) error {

	// Normalize slashes
	basePath = filepath.FromSlash(basePath)

	carefulSave := false
	if len(saveOptions) > 0 {
		for _, saveOption := range saveOptions {
			if saveOption == SaveCareful {
				carefulSave = true
			}
		}
	}

	// Get filepath from interface
	path := filepath.Join(basePath, dataUnit.Filepath())

	var bytes []byte
	var err error

	fpathLower := path[len(path)-5:] // Only need to compare the last 5 characters

	// Use filepath to determine file marshal type
	if fpathLower == `.yaml` {
		bytes, err = yaml.Marshal(dataUnit)
	} else if fpathLower == `.json` {
		bytes, err = json.Marshal(dataUnit)
	} else {
		return errors.New(fmt.Sprint(`SaveFlatFile`, `basePath`, basePath, `type`, fmt.Sprintf(`%T`, *new(T)), `path`, path, `err`, `unsupported file type`))
	}

	if err != nil {
		return errors.New(fmt.Sprint(`SaveFlatFile`, `basePath`, basePath, `type`, fmt.Sprintf(`%T`, *new(T)), `path`, path, `err`, err))
	}

	saveFilePath := path
	if carefulSave { // careful save first saves a {filename}.new file
		saveFilePath += `.new`
	}

	//
	// write to .new suffix in case of power loss etc.
	//
	if err := os.WriteFile(saveFilePath, bytes, 0777); err != nil {
		return errors.New(fmt.Sprint(`SaveAllFlatFiles`, `basePath`, basePath, `type`, fmt.Sprintf(`%T`, *new(T)), `path`, path, `err`, err))
	}

	if carefulSave {
		//
		// Once the file is written, rename it to remove the .new suffix and overwrite the old file
		//
		if err := os.Rename(saveFilePath, path); err != nil {
			return errors.New(fmt.Sprint(`SaveAllFlatFiles`, `basePath`, basePath, `type`, fmt.Sprintf(`%T`, *new(T)), `path`, path, `err`, err))
		}
	}

	return nil
}

// Returns the number of files saved and error
func SaveAllFlatFiles[K comparable, T Loadable[K]](basePath string, data map[K]T, saveOptions ...SaveOption) (int, error) {

	// Normalize slashes
	basePath = filepath.FromSlash(basePath)

	var saveCt int32

	workerCt := runtime.GOMAXPROCS(0)

	var wg sync.WaitGroup
	tData := make(chan T, 1)

	carefulSave := false
	if len(saveOptions) > 0 {
		for _, saveOption := range saveOptions {
			if saveOption == SaveCareful {
				carefulSave = true
			}
		}
	}

	// Spin up workers
	for i := 0; i < workerCt; i++ {

		wg.Add(1)

		go func(dataIn chan T, waitGroup *sync.WaitGroup) {
			defer waitGroup.Done()

			var bytes []byte
			var err error
			var ct int32 = 0

			for dataUnit := range dataIn {

				// Get filepath from interface
				path := filepath.Join(basePath, dataUnit.Filepath())
				fpathLower := path[len(path)-5:] // Only need to compare the last 5 characters

				// Use filepath to determine file marshal type
				if fpathLower == `.yaml` {
					bytes, err = yaml.Marshal(dataUnit)
				} else if fpathLower == `.json` {
					bytes, err = json.Marshal(dataUnit)
				} else {
					panic(fmt.Sprint(`SaveAllFlatFiles`, `basePath`, basePath, `type`, fmt.Sprintf(`%T`, *new(T)), `path`, path, `err`, `unsupported file type`))
				}

				if err != nil {
					panic(fmt.Sprint(`SaveAllFlatFiles`, `basePath`, basePath, `type`, fmt.Sprintf(`%T`, *new(T)), `path`, path, `err`, err))
				}

				saveFilePath := path
				if carefulSave { // careful save first saves a {filename}.new file
					saveFilePath += `.new`
				}

				//
				// write to .new suffix in case of power loss etc.
				//
				if err := os.WriteFile(saveFilePath, bytes, 0777); err != nil {
					panic(fmt.Sprint(`SaveAllFlatFiles`, `basePath`, basePath, `type`, fmt.Sprintf(`%T`, *new(T)), `path`, path, `err`, err))
				}

				if carefulSave {
					//
					// Once the file is written, rename it to remove the .new suffix and overwrite the old file
					//
					if err := os.Rename(saveFilePath, path); err != nil {
						panic(fmt.Sprint(`SaveAllFlatFiles`, `basePath`, basePath, `type`, fmt.Sprintf(`%T`, *new(T)), `path`, path, `err`, err))
					}
				}

				// count saves
				ct++
			}

			atomic.AddInt32(&saveCt, ct)

		}(tData, &wg)
	}

	// Feed all of the data to workers
	for _, d := range data {
		tData <- d
	}

	// Close the channel and wait for workers to finish
	close(tData)

	wg.Wait()

	return int(saveCt), nil
}
