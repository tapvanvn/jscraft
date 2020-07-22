package entity

import (
	"log"

	"com.newcontinent-team.jscraft/tokenize"
)

//BuilderContext context for building per task
type BuilderContext struct {
	Context   *CompileContext
	Templates map[string]*tokenize.BaseToken
	FileScope *JSScopeFile
}

//Init init before use
func (ctxBuild *BuilderContext) Init(jsScopeFile *JSScopeFile, compileContext *CompileContext) {
	ctxBuild.Context = compileContext
	ctxBuild.Templates = make(map[string]*tokenize.BaseToken)
	ctxBuild.FileScope = jsScopeFile
}

//AddTemplate template in file
func (ctxBuild *BuilderContext) AddTemplate(name string, token *tokenize.BaseToken) {

	ctxBuild.Templates[name] = token
}

//GetTemplate get template
func (ctxBuild *BuilderContext) GetTemplate(name string) *tokenize.BaseToken {

	for name, _ := range ctxBuild.Templates {
		log.Println("ctxBuild template:" + name)
	}
	if token, ok := ctxBuild.Templates[name]; ok {

		return token
	}
	return nil
}
