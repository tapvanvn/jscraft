package entity

import "com.newcontinent-team.jscraft/tokenize"

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
func (patchContext *PatchContext) AddPatch(patchName string, contentStream tokenize.BaseTokenStream) {

	patchContext.Patches[patchName] = contentStream
}

//GetPatch get patch
func (patchContext *PatchContext) GetPatch(patchName string) *tokenize.BaseTokenStream {

	if stream, ok := patchContext.Patches[patchName]; ok {

		return &stream

	} else if patchContext.Parent != nil {

		return patchContext.Parent.GetPatch(patchName)
	}

	return patchContext.Context.GetGlobalPatch(patchName)
}
