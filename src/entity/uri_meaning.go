package entity

import (
	"errors"

	"newcontinent-team.com/jscraft/tokenize"
)

const (
	TokenURIOperator = iota
	TokenURIWord
)

//URIMeaning meaning processing for target
type URIMeaning struct {
	Stream       tokenize.BaseTokenStream
	Namespace    string
	RelativePath string
}

//URIOperators separator
var URIOperators []rune = []rune(":/?&=#.")

//Init init before use
func (meaning *URIMeaning) Init(uri string) error {

	if len(uri) > 0 {

		var s tokenize.WordTokenStream

		s.Tokenize(uri)

		var curTokenRunes []rune

		var curType int = tokenize.TokenUnknown

		var pass bool = false
		for {
			if s.EOS() {

				break
			}

			var curchar rune = s.ReadCharacter()

			//detect operator
			if !pass && tokenize.IndexOf(URIOperators, curchar) >= 0 {

				//detect number
				if curchar == ':' {
					//if last is word so it is a namespace
					if len(curTokenRunes) > 0 && curType == TokenURIWord {

						meaning.Namespace = string(curTokenRunes)

						if s.ReadCharacter() != '/' || s.ReadCharacter() != '/' {

							return errors.New("bad syntax")
						}
						pass = true

					} else {

						return errors.New("bad syntax")
					}
				} else {

					return errors.New("bad syntax")
				}

				curTokenRunes = make([]rune, 0)

				curTokenRunes = append(curTokenRunes, curchar)

				curType = TokenURIOperator

			} else {
				if curType == TokenURIOperator {

					curType = TokenURIWord

					curTokenRunes = make([]rune, 0)
				}
				curTokenRunes = append(curTokenRunes, curchar)
			}
		}
		meaning.RelativePath = string(curTokenRunes)

		return nil
	}
	return errors.New("bad uri")
}

//GetNextMeaningToken apply inteface Meaning GetNextMeaningToken
func (meaning *URIMeaning) GetNextMeaningToken() *tokenize.BaseToken {

	for {
		if meaning.Stream.EOS() {
			break
		}
		_ = meaning.Stream.ReadToken()
	}
	return nil
}
