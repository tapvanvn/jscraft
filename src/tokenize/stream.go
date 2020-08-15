package tokenize

import (
	"fmt"
	"strings"
)

//TokenStream token stream
type TokenStream struct {
	Tokens []BaseToken
	//Offset int
	//Level  int
}

//TokenStreamIterator struct use to access token stream
type TokenStreamIterator struct {
	Stream *TokenStream
	Offset int
	Level  int
}

//Iterator make iterator of stream
func (stream *TokenStream) Iterator() TokenStreamIterator {

	return TokenStreamIterator{Stream: stream, Offset: 0, Level: 0}
}

//Tokenize tokenize a string
func (stream *TokenStream) Tokenize(content string) {

	runes := []rune(content)

	for _, rune := range runes {

		token := BaseToken{Content: string(rune)}

		stream.AddToken(token)
	}
}

//AddToken add token to stream
func (stream *TokenStream) AddToken(token BaseToken) {

	stream.Tokens = append(stream.Tokens, token)
}

//AddTokenFromString split string to character and add each character as a token with type is providing type.
func (stream *TokenStream) AddTokenFromString(tokenType int, str string) {

	for _, r := range []rune(str) {

		stream.AddToken(BaseToken{Type: tokenType, Content: string(r)})
	}
}

//AddTokenByContent add token
func (stream *TokenStream) AddTokenByContent(content []rune, tokenType int) {

	stream.Tokens = append(stream.Tokens, BaseToken{Content: string(content), Type: tokenType})
}

//Debug print debug tree
func (stream *TokenStream) Debug(level int, fnName func(int) string) {

	for _, token := range stream.Tokens {

		trimContent := strings.Trim(token.Content, " \n\r")

		if len(trimContent) > 0 || token.Children.Length() > 0 {

			for i := 0; i <= level; i++ {

				if i == 0 {

					fmt.Printf("|%s ", ColorType(token.Type))

				} else {

					fmt.Print("| ")
				}
			}

			if fnName != nil {

				if len(trimContent) > 0 {

					fmt.Printf("%s", ColorContent(token.Content))

				} else {

					fmt.Print("")
				}
				fmt.Printf("-%s\n", ColorName(fnName(token.Type)))

			} else {

				if len(trimContent) > 0 {

					fmt.Println(token.Content)

				} else {

					fmt.Println("")
				}
			}

		}
		token.Children.Debug(level+1, fnName)
	}
}

//DebugMark debug mark
func (iterator *TokenStreamIterator) DebugMark(level int, mark *Mark, ignores *[]int, fnName func(int) string) {

	length := mark.End - mark.Begin

	iter := 0

	for {
		if length <= 0 || iterator.EOS() {
			break
		}

		token := iterator.GetTokenAt(mark.Begin + iter)
		fmt.Printf("%s", ColorOffset(mark.Begin+iter))
		if token != nil {

			for i := 0; i <= level; i++ {

				if i == 0 {

					fmt.Printf("|%s ", ColorType(token.Type))

				} else {

					fmt.Print("| ")
				}
			}

			if !isIgnoreInMark(mark.Begin+iter, ignores) {

				trimContent := strings.Trim(token.Content, " \n\r")

				if len(trimContent) > 0 {

					fmt.Printf("%s", ColorContent(token.Content))

				} else {

					fmt.Print("")
				}

				fmt.Printf("-%s\n", ColorName(fnName(token.Type)))

			} else {

				fmt.Printf("%s", ColorIgnore())
			}

		} else {

			fmt.Printf("%s", "nil")
		}

		fmt.Println("")

		length--

		iter++
	}
}

//GetToken read token but not move pointer
func (iterator *TokenStreamIterator) GetToken() *BaseToken {

	if iterator.Offset <= len(iterator.Stream.Tokens)-1 {

		off := iterator.Offset

		return &iterator.Stream.Tokens[off]
	}
	return nil
}

//GetTokenIter get token at (offset + iterator) position
func (iterator *TokenStreamIterator) GetTokenIter(iter int) *BaseToken {

	if iterator.Offset+iter <= len(iterator.Stream.Tokens)-1 {

		off := iterator.Offset + iter

		return &iterator.Stream.Tokens[off]
	}
	return nil
}

//GetTokenAt get token at offset
func (stream *TokenStream) GetTokenAt(offset int) *BaseToken {

	if offset <= len(stream.Tokens)-1 {

		return &stream.Tokens[offset]
	}
	return nil
}

//GetTokenAt get token at offset
func (iterator *TokenStreamIterator) GetTokenAt(offset int) *BaseToken {

	if offset <= len(iterator.Stream.Tokens)-1 {

		return &iterator.Stream.Tokens[offset]
	}
	return nil
}

//FindPattern search pattern
func (iterator *TokenStreamIterator) FindPattern(patterns []Pattern, stopWhenFound bool, phraseBreak int, isIgnore func(int) bool, fnName func(int) string) []Mark {

	marks := []Mark{}

	log := &Log{}

	//defer log.Print()

	for _, pattern := range patterns {

		iter := 0

		iterToken := 0

		traceIterToken := -1

		patternTokenNum := len(pattern.Struct)

		ignores := []int{}

		children := []*Mark{}

		var patternToken PatternToken

		var childMark *Mark = nil

		for {
			if iterToken >= patternTokenNum {

				mark := Mark{Type: pattern.Type, Begin: iterator.Offset, End: iterator.Offset + iter, Ignores: ignores, Children: children}

				marks = append(marks, mark)

				log.Append(fmt.Sprintf("=>[%s] \n", ColorSuccess()))

				if stopWhenFound {

					return marks
				}
				break
			}
			if iterToken > traceIterToken {

				traceIterToken = iterToken

				patternToken = pattern.Struct[iterToken]

				childMark = &Mark{
					Type:             patternToken.Type,
					CanNested:        patternToken.CanNested,
					IsIgnoreInResult: patternToken.IsIgnoreInResult,
					IsTokenStream:    patternToken.IsPhraseUntil,
				}
				if patternToken.ExportType > 0 {

					childMark.Type = patternToken.ExportType
				}

				children = append(children, childMark)

				childMark.Begin = iterator.Offset + iter

				log.Append(fmt.Sprintf("\n\t[%s %s] %s %t", ColorType(patternToken.Type), ColorName(fnName(patternToken.Type)), ColorContent(patternToken.Content), patternToken.IsPhraseUntil))
			}
			var match bool = true

			var moveIter int = 0

			nextToken := iterator.GetTokenIter(iter)

			if nextToken == nil {
				break
			}

			if nextToken.Type == phraseBreak || isIgnore(nextToken.Type) {

				if pattern.IsRemoveGlobalIgnore || patternToken.IsIgnoreInResult {

					ignores = append(ignores, iterator.Offset+iter)
				}
				iter++

				log.Append(fmt.Sprintf("\n"))

				continue
			}
			if patternToken.Content != "" {

				var currToken = iterator.GetTokenIter(iter)

				if currToken == nil || currToken.Content != patternToken.Content {

					match = false

					log.Append(fmt.Sprintf("=>[%s %s %s]", ColorFail(), ColorType(currToken.Type), ColorContent(currToken.Content)))
				}
				if patternToken.IsIgnoreInResult {

					ignores = append(ignores, iterator.Offset+iter+moveIter)
				}

				childMark.Begin = iterator.Offset + iter

				moveIter = 1

			} else if patternToken.Type > 0 {

				var currToken = iterator.GetTokenIter(iter)

				if currToken == nil || (currToken.Type != phraseBreak && currToken.Type != patternToken.Type) {

					match = false

					log.Append(fmt.Sprintf("=>[%s %s %s]", ColorFail(), ColorType(currToken.Type), ColorContent(currToken.Content)))
				}

				if patternToken.IsIgnoreInResult {

					ignores = append(ignores, iterator.Offset+iter+moveIter)
				}

				if currToken.Type == patternToken.Type {

					childMark.Begin = iterator.Offset + iter
				}

				moveIter = 1

			} else if patternToken.IsPhraseUntil {

				isWordFound := false

				for {
					var currToken = iterator.GetTokenIter(iter + moveIter)

					if currToken == nil {

						match = false

						log.Append(fmt.Sprintf("=>[%s]", ColorFail()))

						break
					}
					if isIgnore(currToken.Type) {

						if pattern.IsRemoveGlobalIgnore || patternToken.IsIgnoreInResult {

							ignores = append(ignores, iterator.Offset+iter+moveIter)
						}
						moveIter++

						continue
					}
					if currToken.Type == phraseBreak && isWordFound {

						if pattern.IsRemoveGlobalIgnore || patternToken.IsIgnoreInResult {

							ignores = append(ignores, iterator.Offset+iter+moveIter)
						}
						moveIter++

						break

					} else if currToken.Type != phraseBreak && len(currToken.Content) > 0 {

						isWordFound = true
					}

					if patternToken.IsIgnoreInResult {

						ignores = append(ignores, iterator.Offset+iter+moveIter)
					}

					moveIter++
				}
			}
			if !match {

				break
			}

			iter += moveIter

			childMark.End = iterator.Offset + iter

			iterToken++
			log.Append(fmt.Sprintf("\n"))
		}
	}
	return marks
}

func isIgnoreInMark(iterator int, ignores *[]int) bool {

	for _, i := range *ignores {

		if i == iterator {

			return true
		}
	}
	return false
}

//GetMaskedToken get token from mask
func (iterator *TokenStreamIterator) GetMaskedToken(mark *Mark, ignores *[]int) *BaseToken {

	if mark.IsTokenStream {

		token := BaseToken{Type: mark.Type}

		len := mark.End - mark.Begin

		iter := 0

		for {
			if len <= 0 || iterator.EOS() {

				break
			}
			nextToken := iterator.GetTokenAt(mark.Begin + iter)

			if !isIgnoreInMark(mark.Begin+iter, ignores) {

				token.Children.AddToken(*nextToken)

			}
			len--

			iter++
		}

		return &token

	} else {

		len := mark.End - mark.Begin

		iter := 0

		for {
			if len <= 0 || iterator.EOS() {

				break
			}
			nextToken := iterator.GetTokenAt(mark.Begin + iter)

			if !isIgnoreInMark(mark.Begin+iter, ignores) {

				return nextToken

			}
			len--

			iter++
		}
	}
	return nil
}

//ReadToken read token
func (iterator *TokenStreamIterator) ReadToken() *BaseToken {

	if iterator.Offset <= len(iterator.Stream.Tokens)-1 {

		off := iterator.Offset

		iterator.Offset++

		return &iterator.Stream.Tokens[off]
	}
	return nil
}

//ResetToBegin reset to begin
func (iterator *TokenStreamIterator) ResetToBegin() {

	iterator.Offset = 0
}

//EOS is end of stream
func (iterator *TokenStreamIterator) EOS() bool {

	return iterator.Offset >= len(iterator.Stream.Tokens)
}

//Length get len of stream
func (stream *TokenStream) Length() int {

	return len(stream.Tokens)
}

//ConcatStringContent concat content of tokens
func (stream *TokenStream) ConcatStringContent() string {

	var iterator = stream.Iterator()

	iterator.ResetToBegin()

	content := ""

	for {
		if iterator.EOS() {

			break
		}
		token := iterator.ReadToken()

		content += string(token.Content)
	}

	return content
}

//ToArray get array of tokens
func (stream *TokenStream) ToArray() []BaseToken {

	var rs []BaseToken

	var iterator = stream.Iterator()

	iterator.ResetToBegin()

	for {
		if iterator.EOS() {

			break
		}
		token := iterator.ReadToken()

		rs = append(rs, *token)
	}

	return rs
}

//ReadFirstTokenType read first token of type
func (stream *TokenStream) ReadFirstTokenType(tokenType int) *BaseToken {

	var iterator = stream.Iterator()

	iterator.ResetToBegin()

	for {
		if iterator.EOS() {

			break
		}
		token := iterator.ReadToken()

		if token.Type == tokenType {

			return token
		}
	}

	return nil
}

//ReadNextTokenType read from current position to next match of token type
func (iterator *TokenStreamIterator) ReadNextTokenType(tokenType int) *BaseToken {

	for {
		if iterator.EOS() {

			break
		}
		token := iterator.ReadToken()

		if token.Type == tokenType {

			return token
		}
	}
	return nil
}
