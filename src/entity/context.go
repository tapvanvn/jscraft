package entity

import (
	"fmt"
	"strconv"
	"strings"

	"sync"

	"newcontinent-team.com/jscraft/tokenize"
)

//Patches patch type
type Patches = map[string]tokenize.BaseToken

//CompileContext conntext for compiles work
type CompileContext struct {
	TemplateDir string

	LayoutDir string

	WorkDir string

	RequireProvider *(chan *JSScopeFile)

	//cache provider
	cacheProvider map[string]*JSScopeFile

	cache_provider sync.Mutex

	//patch
	patches Patches

	filePatches map[string]Patches

	file_patch_mux sync.Mutex

	IsDebug bool

	//url
	cacheURI map[string]string

	cache_uri_mux sync.Mutex
}

//Init init context
func (context *CompileContext) Init() {

	context.patches = make(Patches, 0)

	context.filePatches = make(map[string]Patches)

	context.cacheURI = make(map[string]string, 0)

	context.cacheProvider = make(map[string]*JSScopeFile)
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

	context.cache_uri_mux.Lock()

	cache, ok := context.cacheURI[uri]

	context.cache_uri_mux.Unlock()

	if ok {

		return cache, nil
	}
	var meaning URIMeaning

	err := meaning.Init(uri)

	if err != nil {

		return "", err
	}

	path := context.GetPathForNamespace(meaning.Namespace) + "/" + meaning.RelativePath

	context.cache_uri_mux.Lock()

	context.cacheURI[uri] = path

	context.cache_uri_mux.Unlock()

	return path, nil
}

//RequireJSFile ...
func (context *CompileContext) RequireJSFile(file string) *JSScopeFile {

	context.cache_provider.Lock()

	jsScopeFile, ok := context.cacheProvider[file]

	context.cache_provider.Unlock()

	if ok {

		return jsScopeFile
	}

	scopeFile := JSScopeFile{State: FileStateWaiting}

	scopeFile.Init()

	scopeFile.FilePath = file

	context.cache_provider.Lock()

	context.cacheProvider[file] = &scopeFile

	context.cache_provider.Unlock()

	filePointer := &scopeFile

	go context.require(filePointer)

	return filePointer
}

func (context *CompileContext) require(file *JSScopeFile) {

	(*context.RequireProvider) <- file
}

//MakeBuildContext make builder context
func (context *CompileContext) MakeBuildContext(fileScope *JSScopeFile) *BuilderContext {

	if !fileScope.IsReady() {

		return nil
	}

	table := CreateRequireTable(fileScope.FilePath)

	fileScope.FetchRequire(table)

	builderContext := BuilderContext{}

	builderContext.Init(fileScope, context)

	for requireFile, _ := range table.Files {

		context.cache_provider.Lock()

		fileScope, _ := context.cacheProvider[requireFile]

		context.cache_provider.Unlock()

		fileScope.FetchTemplate(&builderContext)
	}

	return &builderContext
}

//MakePatchContext make builder context
func (context *CompileContext) MakePatchContext(fileScope *JSScopeFile) *PatchContext {

	if !fileScope.IsReady() {

		return nil
	}

	patchContext := PatchContext{}

	patchContext.Init(nil, context)

	context.file_patch_mux.Lock()

	if patches, ok := context.filePatches[fileScope.FilePath]; ok {

		for patchName, patch := range patches {

			patchContext.AddPatch(patchName, patch)
		}
	}

	context.file_patch_mux.Unlock()

	return &patchContext
}

//AddGlobalPatch ...
func (context *CompileContext) AddGlobalPatch(name string, token tokenize.BaseToken) {

	context.file_patch_mux.Lock()

	context.patches[name] = token

	context.file_patch_mux.Unlock()
}

//AddPatch add a patch
func (context *CompileContext) AddPatch(file string, name string, token tokenize.BaseToken) {

	context.file_patch_mux.Lock()

	if _, ok := context.filePatches[file]; !ok {

		context.filePatches[file] = make(Patches)
	}

	context.filePatches[file][name] = token

	context.patches[name] = token

	context.file_patch_mux.Unlock()

}

//GetGlobalPatch get patch in global patch
func (context *CompileContext) GetGlobalPatch(name string) *tokenize.BaseToken {

	context.file_patch_mux.Lock()

	if token, ok := context.patches[name]; ok {

		context.file_patch_mux.Unlock()

		return &token
	}

	context.file_patch_mux.Unlock()

	return nil
}

//GetPatch get patch via name
func (context *CompileContext) GetPatch(file string, name string) *tokenize.BaseToken {

	context.file_patch_mux.Lock()

	if patches, ok := context.filePatches[file]; ok {

		if stream, ok := patches[name]; ok {

			context.file_patch_mux.Unlock()

			return &stream
		}
	}
	context.file_patch_mux.Unlock()

	return context.GetGlobalPatch(name)

}

//Debug debug
func (context *CompileContext) Debug() {

	fmt.Println("--------------begin template-----------")
	for _, scope := range context.cacheProvider {

		fmt.Println("file:" + scope.FilePath)
		fmt.Println("state:" + strconv.Itoa(scope.State))
		for name := range scope.Templates {
			fmt.Println("\ttemp:" + name)
		}
	}

	fmt.Println("--------------end template-----------")
}

//DebugDependence print debug dependence of filepath
func (context *CompileContext) DebugDependence(fileScope *JSScopeFile) {
	fmt.Println("-----begin dependency debug----")
	defer fmt.Println("----end dependency debug----")

	/*fmt.Println(fileScope.FilePath)

	table, ok := context.cacheRequireTable[fileScope.FilePath]

	if !ok {

		fmt.Println("not found require table")
	}

	for _, requireFile := range table.Files {

		fmt.Println("require:" + requireFile)

		scope, ok := context.cacheProvider[requireFile]

		if !ok {

			fmt.Println("ERR2:" + requireFile)
			//todo: error
		}

		if scope.State != FileStateLoaded {

			fmt.Println("file not ready:" + scope.FilePath)
		}
	}*/
}
