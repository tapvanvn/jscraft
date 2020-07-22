package entity

import (
	"errors"
	"strings"

	"com.newcontinent-team.jscraft/tokenize"
	"com.newcontinent-team.jscraft/tokenize/js"
)

//JSMeaning JSMeaning
type JSMeaning struct {
	Stream tokenize.BaseTokenStream

	Context *CompileContext
}

var jsOperators []rune = []rune("#%^&*-+/!<>=?:@\"' \\;\r\n\t{}[](),.|")

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

				if curType != js.TokenJSUnknown {

					meaning.Stream.AddTokenByContent(curTokenRunes, curType)
				}

				curTokenRunes = make([]rune, 0)

				curTokenRunes = append(curTokenRunes, curchar)

				curType = js.TokenJSOperator

			} else {

				if curType == js.TokenJSOperator {

					meaning.Stream.AddTokenByContent(curTokenRunes, curType)

					curTokenRunes = make([]rune, 0)
				}

				curType = js.TokenJSWord
				curTokenRunes = append(curTokenRunes, curchar)
			}
		}
		meaning.Stream.AddTokenByContent(curTokenRunes, curType)

		return nil
	}
	return errors.New("bad content")
}

//GetNextMeaningToken apply inteface Meaning GetNextMeaningToken
func (meaning *JSMeaning) GetNextMeaningToken() *tokenize.BaseToken {

	var token *tokenize.BaseToken = nil

	if meaning.Stream.EOS() {

		return nil
	}

	token = meaning.Stream.ReadToken()

	lower := strings.ToLower(token.GetContent())

	if lower == "{" {

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

	} else if lower == "=" {

		nextToken := meaning.Stream.GetToken()
		if nextToken != nil {
			if nextToken.Content == ">" {
				tmpToken := tokenize.BaseToken{Content: "=>", Type: js.TokenJSRightArrow}
				_ = meaning.Stream.ReadToken()
				return &tmpToken
			}
		}
		tmpToken := tokenize.BaseToken{Content: "=", Type: js.TokenJSAssign}
		return &tmpToken

	} else if lower == "/" {

		nextToken := meaning.Stream.GetToken()

		if nextToken != nil {

			if nextToken.Content == "/" {

				tmpToken := tokenize.BaseToken{Content: "//", Type: js.TokenJSLineComment}

				_ = meaning.Stream.ReadToken()

				meaning.continueReadLineComment(&tmpToken)

				return &tmpToken

			} else if nextToken.Content == "*" {

				tmpToken := tokenize.BaseToken{Content: "/*", Type: js.TokenJSBlockComment}

				_ = meaning.Stream.ReadToken()

				meaning.continueReadBlockComment(&tmpToken)

				return &tmpToken
			} else {
				if meaning.testRegex() {
					tmpToken := tokenize.BaseToken{Content: "/", Type: js.TokenJSRegex}
					tmpToken.Children.AddToken(tokenize.BaseToken{Type: js.TokenJSWord, Content: "/"})

					meaning.continueReadRegex(&tmpToken)

					return &tmpToken
				} else {
					return &tokenize.BaseToken{Content: "/", Type: js.TokenJSOperator}
				}
			}
		}

	} else if lower == " " || lower == "\t" {

		return meaning.GetNextMeaningToken()

	} else if lower == ";" || lower == "\n" || lower == "\r" {

		token.Type = js.TokenJSPhraseBreak
	}

	return token
}

func (meaning *JSMeaning) testRegex() bool {

	var i = meaning.Stream.Offset + 1
	for {
		tmpToken := meaning.Stream.GetTokenAt(i)

		if tmpToken == nil {
			return false
		}

		if tmpToken.Content == "/" {

			testToken := meaning.Stream.GetTokenAt(i + 1)
			if testToken.Content == "i" || testToken.Content == "m" || testToken.Content == "g" {
				return true

			} else {
				return false
			}
		}
		i += 1
	}
	return false
}

func (meaning *JSMeaning) continueReadBracket(currToken *tokenize.BaseToken) {

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
}

func (meaning *JSMeaning) continueReadLineComment(currToken *tokenize.BaseToken) {

	for {
		if meaning.Stream.EOS() {

			break
		}

		tmpToken := meaning.Stream.ReadToken()

		if tmpToken.Content == "\n" || tmpToken.Content == "\r" {

			break

		} else {

			currToken.Children.AddToken(*tmpToken)
		}
	}
}

func (meaning *JSMeaning) continueReadBlockComment(currToken *tokenize.BaseToken) {

	for {
		if meaning.Stream.EOS() {

			break
		}
		tmpToken := meaning.Stream.ReadToken()

		if tmpToken.Content == "*" {

			nextToken := meaning.Stream.GetToken()

			if nextToken != nil && nextToken.Content == "/" {

				_ = meaning.Stream.ReadToken()

				return
			}
		} else {

			currToken.Children.AddToken(*tmpToken)
		}
	}
}

func (meaning *JSMeaning) continueReadRegex(currToken *tokenize.BaseToken) {

	//todo: check syntax violence
	var specialCharacter bool = false
	var gotClose bool = false

	for {
		if meaning.Stream.EOS() {

			break
		}

		tmpToken := meaning.Stream.GetToken()

		tmpContent := tmpToken.GetContent()

		if tmpContent == "\\" {

			specialCharacter = !specialCharacter

			currToken.Children.AddToken(*tmpToken)

			_ = meaning.Stream.ReadToken()

		} else if tmpContent == "/" {

			if specialCharacter {

				specialCharacter = false

			} else {

				gotClose = true
			}

			currToken.Children.AddToken(*tmpToken)

			_ = meaning.Stream.ReadToken()

		} else {

			if gotClose && tmpContent != "i" && tmpContent != "m" && tmpContent != "g" {

				break

			} else {

				_ = meaning.Stream.ReadToken()

				specialCharacter = false

				currToken.Children.AddToken(*tmpToken)
			}
		}
	}
}
