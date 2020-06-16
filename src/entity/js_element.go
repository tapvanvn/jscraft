package entity

import (
	"com.newcontinent-team.jscraft/tokenize"
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
