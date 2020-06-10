package entity

import (
	"fmt"
	"strings"

	"com.newcontinent-team.jscraft/tokenize"
)

//ConpileContext global conntext for compiles work
type CompileContext struct {
	TemplateDir       string
	LayoutDir         string
	WorkDir           string
	Global            JSScopeGlobal
	CacheProvider     map[string]*JSScopeFile
	RequireProvider   *(chan *JSScopeFile)
	cacheRequireTable map[string]*[]string
	patches           map[string]tokenize.BaseTokenStream
	IsDebug           bool

	cacheUri map[string]string
}

//Init init context
func (context *CompileContext) Init() {

	context.cacheRequireTable = make(map[string]*[]string)
	context.patches = make(map[string]tokenize.BaseTokenStream, 0)
	context.cacheUri = make(map[string]string, 0)
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

	cache, ok := context.cacheUri[uri]
	if ok {
		return cache, nil
	}

	var meaning URIMeaning
	err := meaning.Init(uri)
	if err != nil {
		return "", err
	}

	path := context.GetPathForNamespace(meaning.Namespace) + "/" + meaning.RelativePath
	context.cacheUri[uri] = path
	return path, nil
}

//RequireJSFile
func (context *CompileContext) RequireJSFile(file string) *JSScopeFile {

	if jsScopeFile, ok := context.CacheProvider[file]; !ok {
		scopeFile := JSScopeFile{}
		scopeFile.Init()
		scopeFile.FilePath = file
		context.CacheProvider[file] = &scopeFile
		fmt.Println("request file:" + file)
		go context.require(&scopeFile)
		return &scopeFile
	} else {
		return jsScopeFile
	}
}

func (context *CompileContext) require(file *JSScopeFile) {
	(*context.RequireProvider) <- file
}

func (context *CompileContext) IsReadyFor(fileScope *JSScopeFile) bool {
	table, ok := context.cacheRequireTable[fileScope.FilePath]
	if !ok {
		fmt.Println("not found table for file in cache, create onne")
		tmpTable := make([]string, 0)
		table = &tmpTable
		context.cacheRequireTable[fileScope.FilePath] = table
	}
	context.fetchRequireTable(fileScope, table)

	if !fileScope.IsLoaded {
		return false
	}

	for _, requireFile := range *table {
		fileScope, ok := context.CacheProvider[requireFile]
		if !ok {
			//todo: error
		}
		if !fileScope.IsLoaded {
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

//AddPatch add a patch
func (context *CompileContext) AddPatch(name string, stream tokenize.BaseTokenStream) {
	fmt.Println("addPatch: " + name)
	context.patches[name] = stream
}

//GetPatch get patch via name
func (context *CompileContext) GetPatch(name string) *tokenize.BaseTokenStream {
	if stream, ok := context.patches[name]; ok {
		return &stream
	}
	return nil
}
