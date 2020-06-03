package entity

import (
	"com.newcontinent-team.jscraft/tokenize"
)

type JSScopeFile struct {
	FilePath string
	IsLoaded bool
	Children []JSScope
	Requires map[string]*JSScopeFile
	Stream   tokenize.BaseTokenStream
}

func (scope *JSScopeFile) Init() {
	scope.Requires = make(map[string]*JSScopeFile, 0)
}

func (scope *JSScopeFile) GetType() int {
	return JSScopeTypeFile
}

func (scope *JSScopeFile) GetChildren() []JSScope {
	return scope.Children

}
func (scope *JSScopeFile) GetContent() string {
	return ""
}
