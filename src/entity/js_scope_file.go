package entity

import (
	"log"
	"strconv"

	"newcontinent-team.com/jscraft/tokenize"
)

const (
	FileStateUnknown = iota
	FileStateError
	FileStateWaiting
	FileStateLoading
	FileStateLoaded
)

//JSScopeFile basicly is a js file
type JSScopeFile struct {
	FilePath string

	State int
	//IsLoaded bool
	Requires map[string]*JSScopeFile

	Stream tokenize.BaseTokenStream

	Templates map[string]*tokenize.BaseToken
}

//Init init scope
func (scope *JSScopeFile) Init() {

	scope.Requires = make(map[string]*JSScopeFile, 0)

	scope.Templates = make(map[string]*tokenize.BaseToken)
}

//AddTemplate template in file
func (scope *JSScopeFile) AddTemplate(name string, token *tokenize.BaseToken) {

	log.Println("add template:" + name + " at file:" + scope.FilePath)

	scope.Templates[name] = token

	//token.Children.Debug(0, js.TokenName)
}

//Debug print common debug
func (scope *JSScopeFile) Debug() {
	log.Println("----begin file debug ---")
	log.Println(scope.FilePath)
	defer log.Println("---end file debug ---")
	for require, fscope := range scope.Requires {
		log.Println("require: " + require + "\n\tstate:" + strconv.Itoa(fscope.State))
		fscope.Debug()
	}
	for name := range scope.Templates {
		log.Println("temp:" + name)
	}
}
