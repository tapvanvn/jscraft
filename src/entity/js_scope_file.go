package entity

import (
	"sync"

	"newcontinent-team.com/jscraft/tokenize"
)

const (
	FileStateUnknown = iota
	FileStateError
	FileStateWaiting
	FileStateLoading
	FileStateLoaded
)

var __file_scope_id = 0

//JSScopeFile basicly is a js file
type JSScopeFile struct {
	FilePath string

	State int
	//IsLoaded bool
	Requires map[string]*JSScopeFile

	RequireCheck map[string]*CheckReady

	Stream tokenize.TokenStream

	Templates map[string]*tokenize.BaseToken

	require_mux sync.Mutex

	require_lock_count int

	template_mux sync.Mutex

	check_ready_mux sync.Mutex

	id int
}

//Init init scope
func (scope *JSScopeFile) Init() {
	scope.id = __file_scope_id
	__file_scope_id++
	scope.Requires = make(map[string]*JSScopeFile, 0)

	scope.Templates = make(map[string]*tokenize.BaseToken)

	scope.RequireCheck = make(map[string]*CheckReady)
}

//AddTemplate template in file
func (scope *JSScopeFile) AddTemplate(name string, token *tokenize.BaseToken) {

	scope.template_mux.Lock()
	scope.Templates[name] = token
	scope.template_mux.Unlock()
	//token.Children.Debug(0, js.TokenName)
}

//FetchTemplate fetch template to build context
func (scope *JSScopeFile) FetchTemplate(buildContext *BuilderContext) {
	scope.template_mux.Lock()
	for templateName, token := range scope.Templates {

		buildContext.AddTemplate(templateName, token)
	}
	scope.template_mux.Unlock()
}

//Debug print common debug
func (scope *JSScopeFile) Debug() {
	/*fmt.Println("----begin file debug ---")
	fmt.Println(scope.FilePath)
	defer fmt.Println("---end file debug ---")
	for require, fscope := range scope.Requires {
		fmt.Println("require: " + require + "\n\tstate:" + strconv.Itoa(fscope.State))
		fscope.Debug()
	}
	for name := range scope.Templates {
		fmt.Println("temp:" + name)
	}*/
}

//AddRequire add require file
func (scope *JSScopeFile) AddRequire(path string, file *JSScopeFile) *CheckReady {

	if path == scope.FilePath {

		return nil
	}
	scope.require_lock_count++

	scope.require_mux.Lock()

	for _, checkReady := range scope.RequireCheck {

		if checkReady.FileCheck.FilePath == path {

			scope.require_mux.Unlock()
			return nil
		}
	}

	scope.Requires[path] = file

	scope.require_lock_count--

	scope.require_mux.Unlock()

	scope.check_ready_mux.Lock()

	checkReady := &CheckReady{Parent: nil, FileCheck: file, IsReady: false}

	scope.RequireCheck[path] = checkReady

	scope.check_ready_mux.Unlock()

	return checkReady
}

//IsReady if all require file is loaded
func (scope *JSScopeFile) IsReady() bool {

	ready := scope.State == FileStateLoaded

	scope.check_ready_mux.Lock()

	for _, checkReady := range scope.RequireCheck {

		if !checkReady.IsReady {

			ready = false
			break
		}
	}

	scope.check_ready_mux.Unlock()

	return ready
}

//FetchRequire fetch require to table
func (scope *JSScopeFile) FetchRequire(table *RequireTable) {

	if scope.IsReady() {

		scope.check_ready_mux.Lock()

		for _, checkReady := range scope.RequireCheck {

			table.AddFile(checkReady.FileCheck.FilePath)

			checkReady.FileCheck.FetchRequire(table)
		}
		scope.check_ready_mux.Unlock()
	}
}
