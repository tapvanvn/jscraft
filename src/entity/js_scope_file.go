package entity

import (
	"com.newcontinent-team.jscraft/tokenize"
)

const (
	FileStateUnknown = iota
	FileStateError
	FileStateWaiting
	FileStateLoading
	FileStateLoaded
)

type JSScopeFile struct {
	FilePath string
	State    int
	//IsLoaded bool
	Requires map[string]*JSScopeFile
	Stream   tokenize.BaseTokenStream
}

func (scope *JSScopeFile) Init() {
	scope.Requires = make(map[string]*JSScopeFile, 0)
}
