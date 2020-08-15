package entity

import (
	"fmt"
	"strings"

	"newcontinent-team.com/jscraft/tokenize"
	"newcontinent-team.com/jscraft/tokenize/js"
)

//JSMeaningHighContext apply sentence patterns
type JSMeaningHighContext struct {
	Stream tokenize.TokenStream

	Iterator tokenize.TokenStreamIterator

	Context *CompileContext
}

var jscraftKeywords = ",require,conflict,fetch,template,build,"

//Init init before using
func (meaning *JSMeaningHighContext) Init(stream tokenize.TokenStream, context *CompileContext) error {

	meaning.Context = context

	meaning.Stream = stream

	meaning.Iterator = stream.Iterator()

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
		if meaning.Iterator.EOS() {

			break
		}
		/*for {
			nextToken := meaning.Iterator.GetToken()

			if nextToken == nil {

				return nil

			} else if nextToken.Type == js.TokenJSPhraseBreak {

				_ = meaning.Iterator.ReadToken()

			} else {

				break
			}
		}*/
		//nextToken := meaning.Iterator.GetToken()
		//fmt.Printf("%5d \033[1;36m%s\033[0m\n", nextToken.Type, nextToken.Content)

		marks := meaning.Iterator.FindPattern(js.Patterns, true, js.TokenJSPhraseBreak, isIgnore, js.TokenName)

		if len(marks) > 0 {

			patternMark := marks[0]

			currToken := tokenize.BaseToken{Type: patternMark.Type}

			for _, m := range patternMark.Children {

				if m.IsIgnoreInResult {

					continue
				}

				childToken := meaning.Iterator.GetMaskedToken(m, &patternMark.Ignores)

				if childToken != nil && m.CanNested {

					children := tokenize.TokenStream{}

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

					meaning.Iterator.DebugMark(0, &patternMark, &patternMark.Ignores, js.TokenName)

					meaning.Iterator.DebugMark(1, m, &patternMark.Ignores, js.TokenName)
				}
			}

			meaning.Iterator.Offset = patternMark.End

			return &currToken

		} else {

			//var currToken = tokenize.BaseToken{Type: js.TokenJSPhrase}

			//meaning.continueReadPhrase(&currToken)

			//return &currToken

			currToken := meaning.Iterator.ReadToken()

			if currToken.Type == js.TokenJSBlock ||
				currToken.Type == js.TokenJSBracket {

				children := tokenize.TokenStream{}

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
		if meaning.Iterator.EOS() {

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

	iterator := craftToken.Children.Iterator()

	firstToken := iterator.ReadToken()

	if firstToken == nil || strings.Index(jscraftKeywords, ","+firstToken.Content+",") == -1 {

		return nil
	}

	jscraft.FunctionName = firstToken.Content

	jscraft.Stream = &tokenize.TokenStream{}

	secondToken := iterator.ReadToken()

	if secondToken.Type != js.TokenJSBracket {

		return nil
	}
	iterator2 := secondToken.Children.Iterator()

	if jscraft.FunctionName == "require" || jscraft.FunctionName == "fetch" {

		stringToken := iterator2.ReadToken()

		if stringToken == nil || stringToken.Type != js.TokenJSString {

			return nil
		}
		iterator3 := stringToken.Children.Iterator()

		for {
			if iterator3.EOS() {

				break
			}
			token := iterator3.ReadToken()

			jscraft.Stream.AddToken(*token)
		}
	} else {

		for {

			if iterator2.EOS() {

				break
			}
			jscraft.Stream.AddToken(*iterator2.ReadToken())
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

	iterator := functionToken.Children.Iterator()

	firstToken := iterator.GetToken()

	if firstToken == nil {

		return nil
	}

	if firstToken.Type != js.TokenJSBracket {

		jsfunc.FunctionName = firstToken.Content

		_ = iterator.ReadToken()
	}

	secondToken := iterator.ReadToken()

	if secondToken == nil || secondToken.Type != js.TokenJSBracket {

		return nil
	}
	jsfunc.Params = *secondToken

	thirdToken := iterator.ReadToken()

	if thirdToken == nil || thirdToken.Type != js.TokenJSBlock {

		return nil
	}

	jsfunc.Body = *thirdToken

	return &jsfunc
}

//GetJSFor get jsfor object from token
func GetJSFor(forToken *tokenize.BaseToken) *JSFor {

	jsfor := JSFor{}

	iterator := forToken.Children.Iterator()

	firstToken := iterator.ReadToken()

	if firstToken.Type != js.TokenJSBracket {

		return nil
	}
	jsfor.Declare = *firstToken

	secondToken := iterator.ReadToken()

	if secondToken.Type != js.TokenJSBlock {

		return nil
	}
	jsfor.Body = *secondToken

	return &jsfor
}

//GetJSWhile get jswhile object from token
func GetJSWhile(whileToken *tokenize.BaseToken) *JSWhile {

	jswhile := JSWhile{}

	iterator := whileToken.Children.Iterator()

	firstToken := iterator.ReadToken()

	if firstToken != nil && firstToken.Type != js.TokenJSBracket {

		return nil
	}
	jswhile.Condition = *firstToken

	secondToken := iterator.GetToken()

	if secondToken == nil {
		return nil
	}
	if secondToken.Type == js.TokenJSBlock {

		jswhile.Body = *secondToken

	} else {

		jswhile.Body = tokenize.BaseToken{Type: js.TokenJSPhrase}

		for {
			if iterator.EOS() {

				break
			}
			jswhile.Body.Children.AddToken(*(iterator.ReadToken()))
		}
	}
	return &jswhile
}

//GetJSDo get jsdo object from token
func GetJSDo(doToken *tokenize.BaseToken) *JSDo {

	jsdo := JSDo{}

	iterator := doToken.Children.Iterator()

	firstToken := iterator.GetToken()

	if firstToken == nil {
		return nil
	}

	if firstToken.Type != js.TokenJSBlock && firstToken.Type != js.TokenJSPhrase {

		return nil
	}

	jsdo.Body = *firstToken

	_ = iterator.ReadToken()

	secondToken := iterator.GetToken()

	if secondToken == nil || secondToken.Type != js.TokenJSBracket {

		return nil
	}

	jsdo.Condition = *secondToken

	return &jsdo
}

func GetJSSwitch(switchToken *tokenize.BaseToken) *JSSwitch {

	jsswitch := JSSwitch{}

	iterator := switchToken.Children.Iterator()

	firstToken := iterator.GetToken()

	if firstToken == nil {

		return nil
	}

	if firstToken.Type != js.TokenJSBracket {

		return nil
	}

	jsswitch.Var = *firstToken

	_ = iterator.ReadToken()

	secondToken := iterator.GetToken()

	if secondToken == nil || secondToken.Type != js.TokenJSBlock {

		return nil
	}

	jsswitch.Body = *secondToken

	return &jsswitch
}
