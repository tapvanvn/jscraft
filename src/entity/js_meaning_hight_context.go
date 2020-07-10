package entity

import (
	"fmt"
	"strings"

	"com.newcontinent-team.jscraft/tokenize"
	"com.newcontinent-team.jscraft/tokenize/js"
)

//JSMeaningHighContext apply sentence patterns
type JSMeaningHighContext struct {
	Stream tokenize.BaseTokenStream

	Context *CompileContext
}

var jscraftKeywords = ",require,conflict,fetch,template,build,"

//Init init before using
func (meaning *JSMeaningHighContext) Init(stream tokenize.BaseTokenStream, context *CompileContext) error {

	meaning.Context = context

	meaning.Stream = stream

	return nil
}

func isIgnore(tokenType int) bool {

	for _, i := range js.Ignores {

		if i == tokenType {

			return true
		}
	}
	return false
}

//GetNextMeaningToken get next meaning token
func (meaning *JSMeaningHighContext) GetNextMeaningToken() *tokenize.BaseToken {

	for {
		if meaning.Stream.EOS() {

			break
		}
		/*for {
			nextToken := meaning.Stream.GetToken()

			if nextToken == nil {

				return nil

			} else if nextToken.Type == js.TokenJSPhraseBreak {

				_ = meaning.Stream.ReadToken()

			} else {

				break
			}
		}*/
		//nextToken := meaning.Stream.GetToken()
		//fmt.Printf("%5d \033[1;36m%s\033[0m\n", nextToken.Type, nextToken.Content)

		marks := meaning.Stream.FindPattern(js.Patterns, true, js.TokenJSPhraseBreak, isIgnore, js.TokenName)

		if len(marks) > 0 {

			patternMark := marks[0]

			currToken := tokenize.BaseToken{Type: patternMark.Type}

			for _, m := range patternMark.Children {

				if m.IsIgnoreInResult {

					continue
				}

				childToken := meaning.Stream.GetMaskedToken(m, &patternMark.Ignores)

				if childToken != nil && m.CanNested {

					children := tokenize.BaseTokenStream{}

					subMeaning := JSMeaningHighContext{}

					subMeaning.Init(childToken.Children, meaning.Context)

					for {

						nestedToken := subMeaning.GetNextMeaningToken()

						if nestedToken == nil {

							break
						}

						children.AddToken(*nestedToken)
					}

					childToken.Children = children

				}

				if childToken != nil {

					currToken.Children.AddToken(*childToken)

				} else {

					fmt.Println("get token by mark fail")

					meaning.Stream.DebugMark(0, &patternMark, &patternMark.Ignores, js.TokenName)

					meaning.Stream.DebugMark(1, m, &patternMark.Ignores, js.TokenName)
				}
			}

			meaning.Stream.Offset = patternMark.End

			return &currToken

		} else {

			//var currToken = tokenize.BaseToken{Type: js.TokenJSPhrase}

			//meaning.continueReadPhrase(&currToken)

			//return &currToken

			currToken := meaning.Stream.ReadToken()

			if currToken.Type == js.TokenJSBlock ||
				currToken.Type == js.TokenJSBracket {

				children := tokenize.BaseTokenStream{}

				subMeaning := JSMeaningHighContext{}

				subMeaning.Init(currToken.Children, meaning.Context)

				for {

					nestedToken := subMeaning.GetNextMeaningToken()

					if nestedToken == nil {

						break
					}

					children.AddToken(*nestedToken)
				}

				currToken.Children = children
			}

			return currToken
		}
	}
	return nil
}

func (meaning *JSMeaningHighContext) continueReadPhrase(currToken *tokenize.BaseToken) {

	for {
		if meaning.Stream.EOS() {

			break
		}
		nextToken := meaning.GetNextMeaningToken()

		if nextToken == nil {
			break
		}

		if nextToken.Type == js.TokenJSPhraseBreak {

			break
		}

		currToken.Children.AddToken(*nextToken)
	}
}

//GetJSCraft get jscraft object from jscraft token
func GetJSCraft(craftToken *tokenize.BaseToken) *JSCraft {

	jscraft := JSCraft{}

	craftToken.Children.ResetToBegin()

	firstToken := craftToken.Children.ReadToken()

	if firstToken == nil || strings.Index(jscraftKeywords, ","+firstToken.Content+",") == -1 {

		return nil
	}

	jscraft.FunctionName = firstToken.Content

	jscraft.Stream = &tokenize.BaseTokenStream{}

	secondToken := craftToken.Children.ReadToken()

	if secondToken.Type != js.TokenJSBracket {

		return nil
	}
	secondToken.Children.ResetToBegin()

	if jscraft.FunctionName == "require" || jscraft.FunctionName == "fetch" {

		stringToken := secondToken.Children.ReadToken()

		if stringToken == nil || stringToken.Type != js.TokenJSString {

			return nil
		}
		stringToken.Children.ResetToBegin()

		for {
			if stringToken.Children.EOS() {

				break
			}
			token := stringToken.Children.ReadToken()

			jscraft.Stream.AddToken(*token)
		}
	} else {

		for {

			if secondToken.Children.EOS() {

				break
			}
			jscraft.Stream.AddToken(*secondToken.Children.ReadToken())
		}
	}

	return &jscraft
}

//GetJSFunction get jsfunction object from token
func GetJSFunction(functionToken *tokenize.BaseToken) *JSFunction {

	if functionToken.Type != js.TokenJSFunction && functionToken.Type != js.TokenJSFunctionLambda {

		return nil
	}
	jsfunc := JSFunction{}

	functionToken.Children.ResetToBegin()

	firstToken := functionToken.Children.GetToken()

	if firstToken == nil {

		return nil
	}

	if firstToken.Type != js.TokenJSBracket {

		jsfunc.FunctionName = firstToken.Content

		_ = functionToken.Children.ReadToken()
	}

	secondToken := functionToken.Children.ReadToken()

	if secondToken == nil || secondToken.Type != js.TokenJSBracket {

		return nil
	}
	jsfunc.Params = *secondToken

	thirdToken := functionToken.Children.ReadToken()

	if thirdToken == nil || thirdToken.Type != js.TokenJSBlock {

		return nil
	}

	jsfunc.Body = *thirdToken

	return &jsfunc
}

//GetJSFor get jsfor object from token
func GetJSFor(forToken *tokenize.BaseToken) *JSFor {

	jsfor := JSFor{}

	forToken.Children.ResetToBegin()

	firstToken := forToken.Children.ReadToken()

	if firstToken.Type != js.TokenJSBracket {

		return nil
	}
	jsfor.Declare = *firstToken

	secondToken := forToken.Children.ReadToken()

	if secondToken.Type != js.TokenJSBlock {

		return nil
	}
	jsfor.Body = *secondToken

	return &jsfor
}

//GetJSWhile get jswhile object from token
func GetJSWhile(whileToken *tokenize.BaseToken) *JSWhile {

	jswhile := JSWhile{}

	whileToken.Children.ResetToBegin()

	firstToken := whileToken.Children.ReadToken()

	if firstToken != nil && firstToken.Type != js.TokenJSBracket {

		return nil
	}
	jswhile.Condition = *firstToken

	secondToken := whileToken.Children.GetToken()

	if secondToken == nil {
		return nil
	}
	if secondToken.Type == js.TokenJSBlock {

		jswhile.Body = *secondToken

	} else {

		jswhile.Body = tokenize.BaseToken{Type: js.TokenJSPhrase}

		for {
			if whileToken.Children.EOS() {

				break
			}
			jswhile.Body.Children.AddToken(*(whileToken.Children.ReadToken()))
		}
	}
	return &jswhile
}

//GetJSDo get jsdo object from token
func GetJSDo(doToken *tokenize.BaseToken) *JSDo {

	jsdo := JSDo{}

	doToken.Children.ResetToBegin()

	firstToken := doToken.Children.GetToken()

	if firstToken == nil {
		return nil
	}

	if firstToken.Type != js.TokenJSBlock && firstToken.Type != js.TokenJSPhrase {

		return nil
	}

	jsdo.Body = *firstToken

	_ = doToken.Children.ReadToken()

	secondToken := doToken.Children.GetToken()

	if secondToken == nil || secondToken.Type != js.TokenJSBracket {

		return nil
	}

	jsdo.Condition = *secondToken

	return &jsdo
}

func GetJSSwitch(switchToken *tokenize.BaseToken) *JSSwitch {

	jsswitch := JSSwitch{}

	switchToken.Children.ResetToBegin()

	firstToken := switchToken.Children.GetToken()

	if firstToken == nil {

		return nil
	}

	if firstToken.Type != js.TokenJSBracket {

		return nil
	}

	jsswitch.Var = *firstToken

	_ = switchToken.Children.ReadToken()

	secondToken := switchToken.Children.GetToken()

	if secondToken == nil || secondToken.Type != js.TokenJSBlock {

		return nil
	}

	jsswitch.Body = *secondToken

	return &jsswitch
}
