package entity

import (
	"fmt"

	"newcontinent-team.com/jscraft/tokenize"
)

var __BuildContextID int = 0

//BuilderContext context for building per task
type BuilderContext struct {
	Context   *CompileContext
	Templates map[string]*tokenize.BaseToken
	FileScope *JSScopeFile
	ID        int
}

//Init init before use
func (ctxBuild *BuilderContext) Init(jsScopeFile *JSScopeFile, compileContext *CompileContext) {

	ctxBuild.ID = __BuildContextID

	__BuildContextID++

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

	if token, ok := ctxBuild.Templates[name]; ok {

		return token
	}

	return nil
}

//Debug debug
func (ctxBuild *BuilderContext) Debug() {

	fmt.Println("---begin ctxBuild---")
	fmt.Println(ctxBuild.FileScope.FilePath)
	for name := range ctxBuild.Templates {
		fmt.Println("temp:" + name)
	}
	fmt.Println("---end ctxBuild---")
}
