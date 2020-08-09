package entity

import (
	"log"
	"strconv"
	"strings"

	"sync"

	"newcontinent-team.com/jscraft/tokenize"
)

//Patches patch type
type Patches = map[string]tokenize.BaseToken

//RequireTable save require files to build a file
type RequireTable struct {
	OwnerFile string
	Files     []string
}

//AddFile add file
func (table *RequireTable) AddFile(filePath string) {

	log.Println("\n" + table.OwnerFile + "\n===>" + filePath + "\n")
	table.Files = append(table.Files, filePath)
}

//HasFile check if added file
func (table *RequireTable) HasFile(filePath string) bool {
	for _, file := range table.Files {
		if file == filePath {
			return true
		}
	}
	return false
}

//CompileContext conntext for compiles work
type CompileContext struct {
	TemplateDir string

	LayoutDir string

	WorkDir string

	CacheProvider map[string]*JSScopeFile

	RequireProvider *(chan *JSScopeFile)

	cacheRequireTable map[string]*RequireTable

	patches Patches

	filePatches map[string]Patches

	//builderPatches map[string]Patches

	IsDebug bool

	cacheURI map[string]string

	mux sync.Mutex
}

//Init init context
func (context *CompileContext) Init() {

	context.cacheRequireTable = make(map[string]*RequireTable)

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

		log.Println("not found table for file in cache, create one: \n\t" + fileScope.FilePath)

		context.mux.Lock()

		context.cacheRequireTable[fileScope.FilePath] = &RequireTable{
			OwnerFile: fileScope.FilePath,
			Files:     make([]string, 0)}

		table, _ = context.cacheRequireTable[fileScope.FilePath]

		context.mux.Unlock()
	}

	context.fetchRequireTable(fileScope, table)

	if fileScope.State != FileStateLoaded {

		return false
	}

	for _, requireFile := range table.Files {

		fileScope, ok := context.CacheProvider[requireFile]

		if !ok {

			log.Println("ERR2:" + requireFile)
			//todo: error
		}

		if fileScope.State != FileStateLoaded {

			return false
		}
	}
	return true
}

//MakeBuildContext make builder context
func (context *CompileContext) MakeBuildContext(fileScope *JSScopeFile) *BuilderContext {

	if !context.IsReadyFor(fileScope) {

		return nil
	}

	builderContext := BuilderContext{}

	builderContext.Init(fileScope, context)

	table, _ := context.cacheRequireTable[fileScope.FilePath]

	context.fetchRequireTable(fileScope, table)

	for _, requireFile := range table.Files {

		fileScope, _ := context.CacheProvider[requireFile]

		for templateName, token := range fileScope.Templates {

			builderContext.AddTemplate(templateName, token)
		}
	}

	return &builderContext
}

func (context *CompileContext) fetchRequireTable(fileScope *JSScopeFile, table *RequireTable) {

	for requireFile, requireFileScope := range fileScope.Requires {

		if !table.HasFile(requireFile) {

			table.AddFile(requireFile)

		}
		context.fetchRequireTable(requireFileScope, table)
	}
}

//MakePatchContext make builder context
func (context *CompileContext) MakePatchContext(fileScope *JSScopeFile) *PatchContext {

	if !context.IsReadyFor(fileScope) {

		return nil
	}

	patchContext := PatchContext{}

	patchContext.Init(nil, context)

	context.mux.Lock()

	if patches, ok := context.filePatches[fileScope.FilePath]; ok {

		for patchName, patch := range patches {

			patchContext.AddPatch(patchName, patch)
		}
	}

	context.mux.Unlock()

	return &patchContext
}

//AddGlobalPatch ...
func (context *CompileContext) AddGlobalPatch(name string, token tokenize.BaseToken) {

	context.mux.Lock()

	context.patches[name] = token

	context.mux.Unlock()
}

//AddPatch add a patch
func (context *CompileContext) AddPatch(file string, name string, token tokenize.BaseToken) {

	context.mux.Lock()

	if _, ok := context.filePatches[file]; !ok {
		context.filePatches[file] = make(Patches)
	}

	context.filePatches[file][name] = token

	context.patches[name] = token

	context.mux.Unlock()

}

//GetGlobalPatch get patch in global patch
func (context *CompileContext) GetGlobalPatch(name string) *tokenize.BaseToken {

	if token, ok := context.patches[name]; ok {

		return &token
	}

	return nil
}

//GetPatch get patch via name
func (context *CompileContext) GetPatch(file string, name string) *tokenize.BaseToken {

	if patches, ok := context.filePatches[file]; ok {
		if stream, ok := patches[name]; ok {
			return &stream
		}
	}

	return context.GetGlobalPatch(name)

}

//Debug debug
func (context *CompileContext) Debug() {

	log.Println("--------------begin template-----------")
	for _, scope := range context.CacheProvider {

		log.Println("file:" + scope.FilePath)
		log.Println("state:" + strconv.Itoa(scope.State))
		for name := range scope.Templates {
			log.Println("\ttemp:" + name)
		}
	}

	log.Println("--------------end template-----------")
}

//DebugDependence print debug dependence of filepath
func (context *CompileContext) DebugDependence(fileScope *JSScopeFile) {
	log.Println("-----begin dependency debug----")
	defer log.Println("----end dependency debug----")

	log.Println(fileScope.FilePath)

	table, ok := context.cacheRequireTable[fileScope.FilePath]

	if !ok {

		log.Println("not found require table")
	}

	for _, requireFile := range table.Files {

		log.Println("require:" + requireFile)

		scope, ok := context.CacheProvider[requireFile]

		if !ok {

			log.Println("ERR2:" + requireFile)
			//todo: error
		}

		if scope.State != FileStateLoaded {

			log.Println("file not ready:" + scope.FilePath)
		}
	}
}
