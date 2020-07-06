package entity

import (
	"fmt"
	"strings"

	"sync"

	"com.newcontinent-team.jscraft/tokenize"
)

type Patches = map[string]tokenize.BaseTokenStream

//CompileContext conntext for compiles work
type CompileContext struct {
	TemplateDir string

	LayoutDir string

	WorkDir string

	CacheProvider map[string]*JSScopeFile

	RequireProvider *(chan *JSScopeFile)

	cacheRequireTable map[string]*[]string

	patches Patches

	filePatches map[string]Patches

	IsDebug bool

	cacheURI map[string]string

	mux sync.Mutex
}

//Init init context
func (context *CompileContext) Init() {

	context.cacheRequireTable = make(map[string]*[]string)

	context.patches = make(Patches, 0)

	context.filePatches = make(map[string]Patches)

	context.cacheURI = make(map[string]string, 0)
}

//GetPathForNamespace get
func (context *CompileContext) GetPathForNamespace(namespace string) string {

	switch strings.ToLower(namespace) {

	case "work":
		return context.WorkDir

	case "layout":
		return context.LayoutDir

	case "template":
		return context.TemplateDir
	}
	return ""
}

//GetPathForURI get string for uri
func (context *CompileContext) GetPathForURI(uri string) (string, error) {

	context.mux.Lock()

	cache, ok := context.cacheURI[uri]

	context.mux.Unlock()

	if ok {

		return cache, nil
	}
	var meaning URIMeaning

	err := meaning.Init(uri)

	if err != nil {

		return "", err
	}
	path := context.GetPathForNamespace(meaning.Namespace) + "/" + meaning.RelativePath

	context.mux.Lock()

	context.cacheURI[uri] = path

	context.mux.Unlock()

	return path, nil
}

//RequireJSFile ...
func (context *CompileContext) RequireJSFile(file string) *JSScopeFile {

	jsScopeFile, ok := context.CacheProvider[file]

	if ok {

		return jsScopeFile
	}

	scopeFile := JSScopeFile{State: FileStateWaiting}

	scopeFile.Init()

	scopeFile.FilePath = file

	context.mux.Lock()

	context.CacheProvider[file] = &scopeFile

	context.mux.Unlock()

	filePointer := &scopeFile

	go context.require(filePointer)

	return filePointer
}

func (context *CompileContext) require(file *JSScopeFile) {

	(*context.RequireProvider) <- file
}

//IsReadyFor check if a file is ready for export
func (context *CompileContext) IsReadyFor(fileScope *JSScopeFile) bool {

	table, ok := context.cacheRequireTable[fileScope.FilePath]

	if !ok {

		fmt.Println("not found table for file in cache, create onne")

		tmpTable := make([]string, 0)

		table = &tmpTable

		context.mux.Lock()

		context.cacheRequireTable[fileScope.FilePath] = table

		context.mux.Unlock()
	}
	context.fetchRequireTable(fileScope, table)

	if fileScope.State != FileStateLoaded {

		return false
	}

	for _, requireFile := range *table {

		fileScope, ok := context.CacheProvider[requireFile]

		if !ok {
			//todo: error
		}

		if fileScope.State != FileStateLoaded {

			return false
		}
	}
	return true
}

func (context *CompileContext) fetchRequireTable(fileScope *JSScopeFile, table *[]string) {

	for requireFile, requireFileScope := range fileScope.Requires {

		found := false

		for _, tableFile := range *table {

			if requireFile == tableFile {

				found = true

				break
			}
		}

		if !found {

			*table = append(*table, requireFile)

			context.fetchRequireTable(requireFileScope, table)
		}
	}
}

//AddGlobalPatch ...
func (context *CompileContext) AddGlobalPatch(name string, stream tokenize.BaseTokenStream) {

	context.mux.Lock()

	context.patches[name] = stream

	context.mux.Unlock()
}

//AddPatch add a patch
func (context *CompileContext) AddPatch(file string, name string, stream tokenize.BaseTokenStream) {

	context.mux.Lock()

	if _, ok := context.filePatches[file]; !ok {
		context.filePatches[file] = make(Patches)
	}

	context.filePatches[file][name] = stream

	context.patches[name] = stream

	context.mux.Unlock()

}

//GetPatch get patch via name
func (context *CompileContext) GetPatch(file string, name string) *tokenize.BaseTokenStream {

	if patches, ok := context.filePatches[file]; ok {
		if stream, ok := patches[name]; ok {
			return &stream
		}
	}

	if stream, ok := context.patches[name]; ok {

		return &stream
	}

	return nil
}
