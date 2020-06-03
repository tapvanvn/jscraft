package entity

type JSScopeLine struct {
	Children []JSScope
}

func (scope *JSScopeLine) GetType() int {
	return JSScopeTypeLine
}

func (scope *JSScopeLine) GetChildren() []JSScope {
	return scope.Children

}
func (scope *JSScopeLine) GetContent() string {
	return ""
}
