package entity

import (
	"errors"
	"strings"

	"com.newcontinent-team.jscraft/tokenize"
	"com.newcontinent-team.jscraft/tokenize/js"
)

const (
	//JSScopeTypeGlobal JSScopeTypeGlobal
	JSScopeTypeGlobal = iota
	//JSScopeTypeFile JSScopeTypeFile
	JSScopeTypeFile
	//JSScopeTypeLine JSScopeTypeLine
	JSScopeTypeLine
	//JSScopeTypeFunction JSScopeTypeFunction
	JSScopeTypeFunction
	//JSScopeTypeFunctionCall JSScopeTypeFunctionCall
	JSScopeTypeFunctionCall
)

//JSScope JSScope
type JSScope interface {
	GetType() int
	GetChildren() []JSScope
	GetContent() string
}

//JSMeaning JSMeaning
type JSMeaning struct {
	Stream  tokenize.BaseTokenStream
	Context *CompileContext
}

var jsOperators []rune = []rune("$#%^&*-+/!<>=?:@\"' \\;\r\n\t{}[](),.|")

//Init a string file
func (meaning *JSMeaning) Init(content string, context *CompileContext) error {
	if len(content) > 0 {
		meaning.Context = context
		var s tokenize.WordTokenStream
		s.Tokenize(content)

		var curTokenRunes []rune
		var curType int = js.TokenJSUnknown

		for {
			if s.EOS() {
				break
			}

			var curchar rune = s.ReadCharacter()

			if tokenize.IndexOf(jsOperators, curchar) >= 0 {

				meaning.Stream.AddTokenByContent(curTokenRunes, curType)
				curTokenRunes = make([]rune, 0)
				curTokenRunes = append(curTokenRunes, curchar)
				curType = js.TokenJSOperator

			} else {

				if curType == js.TokenJSOperator {

					meaning.Stream.AddTokenByContent(curTokenRunes, curType)
					curType = js.TokenJSWord
					curTokenRunes = make([]rune, 0)

				}

				curTokenRunes = append(curTokenRunes, curchar)
			}
		}
		meaning.Stream.AddTokenByContent(curTokenRunes, curType)

		//fmt.Println("length:sdf:" + strconv.Itoa(meaning.Stream.Length()))
		//meaning.Stream.ResetToBegin()
		return nil
	}
	return errors.New("bad content")
}

//GetNextMeaningToken apply inteface Meaning GetNextMeaningToken
func (meaning *JSMeaning) GetNextMeaningToken() *tokenize.BaseToken {
	var token *tokenize.BaseToken = nil
	//fmt.Println("length:" + strconv.Itoa(meaning.Stream.Length()))
	if meaning.Stream.EOS() {
		return nil
	}

	token = meaning.Stream.ReadToken()
	lower := strings.ToLower(token.GetContent())

	if lower == "function" {

		tmpToken := tokenize.BaseToken{Type: js.TokenJSFunction}
		meaning.continueReadFunction(&tmpToken)
		return &tmpToken

	} else if lower == "for" {

		tmpToken := tokenize.BaseToken{Type: js.TokenJSFor}
		meaning.continueReadForLoop(&tmpToken)
		return &tmpToken

	} else if lower == "if" {

		tmpToken := tokenize.BaseToken{Type: js.TokenJSIf}
		meaning.continueReadIf(&tmpToken)
		return &tmpToken

	} else if lower == "var" {

		tmpToken := tokenize.BaseToken{Content: "$", Type: js.TokenJSVariable}
		meaning.continueReadVariable(&tmpToken)
		return &tmpToken

	} else if lower == "jscraft" {

		tmpToken := tokenize.BaseToken{Type: js.TokenJSCraft}
		meaning.continueReadCraft(&tmpToken)
		return &tmpToken

	} else if lower == "{" {

		tmpToken := tokenize.BaseToken{Content: "{", Type: js.TokenJSBlock}
		meaning.continueReadBlock(&tmpToken)
		return &tmpToken

	} else if lower == "'" {

		tmpToken := tokenize.BaseToken{Content: "'", Type: js.TokenJSString}
		meaning.continueReadString(&tmpToken)
		return &tmpToken

	} else if lower == "\"" {

		tmpToken := tokenize.BaseToken{Content: "\"", Type: js.TokenJSString}
		meaning.continueReadString(&tmpToken)
		return &tmpToken

	} else if lower == "(" {

		tmpToken := tokenize.BaseToken{Content: "(", Type: js.TokenJSBracket}
		meaning.continueReadBracket(&tmpToken)
		return &tmpToken

	} else if lower == "[" {

		tmpToken := tokenize.BaseToken{Content: "[", Type: js.TokenJSBracketSquare}
		meaning.continueReadBracketSquare(&tmpToken)
		return &tmpToken
	}

	return token
}

func (meaning *JSMeaning) continueReadFunction(currToken *tokenize.BaseToken) {

	for {
		if meaning.Stream.EOS() {
			break
		}

		tmpToken := meaning.GetNextMeaningToken()
		tmpContent := tmpToken.GetContent()

		if tmpToken.Type == js.TokenJSBracket {
			//fmt.Println("function param")

			paramToken := tokenize.BaseToken{Type: js.TokenJSFunctionParam}

			tmpStream := tmpToken.Children

			for {
				if tmpStream.EOS() {
					break
				}

				varToken := tmpStream.ReadToken()

				if varToken.Type == js.TokenJSVariable {
					paramToken.Children.AddToken(*varToken)
				} else if varToken.Type == js.TokenJSWord {
					paramToken.Children.AddToken(*varToken)
				}
			}

			currToken.Children.AddToken(paramToken)

		} else if tmpToken.Type == js.TokenJSBlock {

			currToken.Children.AddToken(*tmpToken)
			break

		} else if len(tmpContent) > 0 {
			name := strings.Trim(tmpContent, " \n\r")
			if len(name) > 0 {
				//fmt.Println("function name" + name + ":" + strconv.Itoa(len(name)))
				currToken.Content = name
			}
		}
	}
}

func (meaning *JSMeaning) continueReadBracket(currToken *tokenize.BaseToken) {

	//fmt.Println("begin block")
	for {
		if meaning.Stream.EOS() {
			break
		}
		tmpToken := meaning.GetNextMeaningToken()

		tmpContent := tmpToken.GetContent()

		if tmpContent == ")" {
			break
		} else {
			currToken.Children.AddToken(*tmpToken)
		}
	}
	//fmt.Println("end block")

}

func (meaning *JSMeaning) continueReadBracketSquare(currToken *tokenize.BaseToken) {

	for {
		if meaning.Stream.EOS() {
			break
		}
		tmpToken := meaning.GetNextMeaningToken()

		tmpContent := tmpToken.GetContent()

		if tmpContent == "]" {
			break
		} else {
			currToken.Children.AddToken(*tmpToken)
		}
	}
}

func (meaning *JSMeaning) continueReadBlock(currToken *tokenize.BaseToken) {

	for {
		if meaning.Stream.EOS() {
			break
		}
		tmpToken := meaning.GetNextMeaningToken()

		tmpContent := tmpToken.GetContent()

		if tmpContent == "}" {
			break
		} else {
			currToken.Children.AddToken(*tmpToken)
		}
	}
}

func (meaning *JSMeaning) continueReadString(currToken *tokenize.BaseToken) {

	//fmt.Println("begin string")

	var specialCharacter bool = false
	curContent := currToken.GetContent()
	for {
		if meaning.Stream.EOS() {
			break
		}
		tmpToken := meaning.Stream.ReadToken()
		tmpContent := tmpToken.GetContent()

		if tmpContent == "\\" {

			specialCharacter = !specialCharacter
			currToken.Children.AddToken(*tmpToken)

		} else if tmpContent == curContent {

			if specialCharacter {
				specialCharacter = false
				currToken.Children.AddToken(*tmpToken)
			} else {
				break
			}

		} else {
			specialCharacter = false
			currToken.Children.AddToken(*tmpToken)
		}
	}
	//fmt.Println("end string")
}

func (meaning *JSMeaning) continueReadForLoop(currToken *tokenize.BaseToken) {

	for {
		if meaning.Stream.EOS() {
			break
		}
		tmpToken := meaning.GetNextMeaningToken()

		currToken.Children.AddToken(*tmpToken)

		if tmpToken.Type == js.TokenJSBlock {
			break
		}
	}
}

func (meaning *JSMeaning) continueReadIf(currToken *tokenize.BaseToken) {

	for {
		if meaning.Stream.EOS() {
			break
		}
		tmpToken := meaning.GetNextMeaningToken()
		currToken.Children.AddToken(*tmpToken)

		if tmpToken.Type == js.TokenJSBlock || tmpToken.Content == ";" {
			break
		}
	}
}

func (meaning *JSMeaning) continueReadVariable(currToken *tokenize.BaseToken) {
	for {
		if meaning.Stream.EOS() {
			break
		}
		tmpToken := meaning.GetNextMeaningToken()

		currToken.Children.AddToken(*tmpToken)
		tmpContent := tmpToken.GetContent()
		if tmpContent == "\n" {
			tmpToken.Content = ";"
			break
		} else if tmpContent == ";" {
			break
		}
	}
}

func (meaning *JSMeaning) continueReadCraft(currToken *tokenize.BaseToken) {
	//fmt.Println("begin craft")
	for {
		if meaning.Stream.EOS() {
			break
		}
		tmpToken := meaning.GetNextMeaningToken()

		tmpContent := tmpToken.Content
		//fmt.Println("content:" + tmpContent)

		if tmpContent == "require" {

			currToken.Content = "require"

		} else if tmpContent == "conflict" {

			currToken.Content = "conflict"

		} else if tmpContent == "fetch" {

			currToken.Content = "fetch"

		} else if tmpContent == "(" {

			if tmpToken.Type == js.TokenJSBracket {
				currToken.Children.AddToken(*tmpToken)
			} else {
				tokenBracket := tokenize.BaseToken{Content: "(", Type: js.TokenJSBracket}
				meaning.continueReadBracket(&tokenBracket)
			}
			break
		} else if tmpContent == "\n" || tmpContent == ";" || tmpContent == "\r" {
			break
		} else if tmpContent == "." {
			continue
		} else {
			//syntax error
			//fmt.Println("unexpected:" + tmpContent)
		}
	}
	//fmt.Println("end craft")
}
