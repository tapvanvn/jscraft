package entity

import (
	"com.newcontinent-team.jscraft/tokenize"
)

//JSCraft infomation about scraft call
type JSCraft struct {
	FunctionName string

	Stream *tokenize.BaseTokenStream
}
