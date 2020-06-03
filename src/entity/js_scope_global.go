package entity

type JSScopeGlobal struct {
	Children  []JSScope
	CacheLoad map[string]*JSScopeFile
}

func (scope *JSScopeGlobal) GetType() int {
	return JSScopeTypeGlobal
}

func (scope *JSScopeGlobal) GetChildren() []JSScope {
	return scope.Children

}

func (scope *JSScopeGlobal) GetContent() string {
	return ""
}
