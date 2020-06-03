package entity

import (
	"errors"

	"com.newcontinent-team.jscraft/tokenize"
)

const (
	TokenURIOperator = iota
	TokenURIWord
)

//TargetMeaning meaning processing for target
type URIMeaning struct {
	Stream       tokenize.BaseTokenStream
	Namespace    string
	RelativePath string
}

var URIOperators []rune = []rune(":/?&=#.")

func (meaning *URIMeaning) Init(uri string) error {

	if len(uri) > 0 {
		//PhpTokenStream token_stream = new PhpTokenStream();
		//StringStream s = new StringStream(nocmmend_content);

		var s tokenize.WordTokenStream
		s.Tokenize(uri)

		//var is_instring bool = false
		//var is_skip bool = false
		//var is_newline bool = true

		var cur_token_runes []rune
		var cur_type int = tokenize.TokenUnknown

		var pass bool = false
		for {

			if s.EOS() {
				break
			}

			//var skip_newline bool = false

			var curchar rune = s.ReadCharacter()

			//detect operator
			if !pass && tokenize.IndexOf(URIOperators, curchar) >= 0 {

				//detect number
				if curchar == ':' {
					//if last is word so it is a namespace
					if len(cur_token_runes) > 0 && cur_type == TokenURIWord {
						meaning.Namespace = string(cur_token_runes)
						if s.ReadCharacter() != '/' || s.ReadCharacter() != '/' {
							return errors.New("bad syntax")
						} else {
							pass = true
						}
					} else {
						return errors.New("bad syntax")
					}
				} else {
					return errors.New("bad syntax")
				}

				//meaning.Stream.AddTokenByContent(cur_token_runes, cur_type)
				cur_token_runes = make([]rune, 0)
				cur_token_runes = append(cur_token_runes, curchar)
				cur_type = TokenURIOperator
			} else {
				if cur_type == TokenURIOperator {
					//meaning.Stream.AddTokenByContent(cur_token_runes, cur_type)
					cur_type = TokenURIWord
					cur_token_runes = make([]rune, 0)
				}
				cur_token_runes = append(cur_token_runes, curchar)
			}
		}
		meaning.RelativePath = string(cur_token_runes)
		return nil
		//meaning.Stream.AddTokenByContent(cur_token_runes, cur_type)
	}
	return errors.New("bad uri")
}

//GetNextMeaningToken apply inteface Meaning GetNextMeaningToken
func (meaning *URIMeaning) GetNextMeaningToken() *tokenize.BaseToken {
	//var token tokenize.Token
	for {
		if meaning.Stream.EOS() {
			break
		}
		_ = meaning.Stream.ReadToken()
	}
	return nil
}
