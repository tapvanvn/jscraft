package entity

import (
	"log"

	"newcontinent-team.com/jscraft/tokenize"
)

//PatchContext patch context
type PatchContext struct {
	Context *CompileContext
	Parent  *PatchContext
	Patches Patches
}

//Init init before use
func (patchContext *PatchContext) Init(parent *PatchContext, compileContext *CompileContext) {

	patchContext.Parent = parent

	patchContext.Context = compileContext

	patchContext.Patches = make(Patches)
}

//AddPatch add patch to patchContext
func (patchContext *PatchContext) AddPatch(patchName string, token tokenize.BaseToken) {

	patchContext.Patches[patchName] = token
}

//GetPatch get patch
func (patchContext *PatchContext) GetPatch(patchName string) *tokenize.BaseToken {

	if token, ok := patchContext.Patches[patchName]; ok {

		return &token

	} else if patchContext.Parent != nil {

		return patchContext.Parent.GetPatch(patchName)
	}

	return patchContext.Context.GetGlobalPatch(patchName)
}

//Debug print debug
func (patchContext *PatchContext) Debug() {
	log.Println("debug----patchContext")
	for name, _ := range patchContext.Patches {

		log.Println("-name-:" + name)

	}
}
