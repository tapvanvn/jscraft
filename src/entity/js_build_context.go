package entity

import (
	"com.newcontinent-team.jscraft/tokenize"
)

//ContextFunction infomation about a function
type ContextFunction struct {
	IsFunction bool

	Name string

	IsLambda bool

	Params []tokenize.BaseToken
}

func analysFunctionContext(token *tokenize.BaseToken) ContextFunction {

	context := ContextFunction{}

	return context
}
