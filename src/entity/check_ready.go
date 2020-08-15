package entity

type CheckReady struct {
	Parent    *CheckReady
	FileCheck *JSScopeFile
	IsReady   bool
}
