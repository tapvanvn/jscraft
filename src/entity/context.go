package entity

import "strings"

//ConpileContext global conntext for compiles work
type CompileContext struct {
	TemplateDir     string
	LayoutDir       string
	WorkDir         string
	Global          JSScopeGlobal
	CacheProvider   map[string]*JSScopeFile
	RequireProvider *(chan *JSScopeFile)
}

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

//RequireJSFile
func (context *CompileContext) RequireJSFile(file string) *JSScopeFile {

	if jsScopeFile, ok := context.CacheProvider[file]; !ok {
		scopeFile := JSScopeFile{}
		scopeFile.FilePath = file
		context.CacheProvider[file] = &scopeFile
		go context.require(&scopeFile)
		return &scopeFile
	} else {
		return jsScopeFile
	}
}

func (context *CompileContext) require(file *JSScopeFile) {
	(*context.RequireProvider) <- file
}
