package entity

type JSScopeFunction struct {
	Children []JSScope
}

func (scope *JSScopeFunction) GetType() int {
	return JSScopeTypeFunction
}

func (scope *JSScopeFunction) GetChildren() []JSScope {
	return scope.Children
}

func (scope *JSScopeFunction) GetContent() string {
	return ""
}
