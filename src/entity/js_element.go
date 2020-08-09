package entity

import (
	"newcontinent-team.com/jscraft/tokenize"
)

//JSFunction infomation of a function
type JSFunction struct {
	FunctionName string

	Params tokenize.BaseToken

	Body tokenize.BaseToken
}

//JSFor infomation of a for loop
type JSFor struct {
	Declare tokenize.BaseToken

	Body tokenize.BaseToken
}

//JSWhile infomation of a while loop
type JSWhile struct {
	Condition tokenize.BaseToken

	Body tokenize.BaseToken
}

//JSDo infomation of a do while loop
type JSDo struct {
	Condition tokenize.BaseToken

	Body tokenize.BaseToken
}

//JSSwitch infomation of a switch
type JSSwitch struct {
	Var  tokenize.BaseToken
	Body tokenize.BaseToken
}
