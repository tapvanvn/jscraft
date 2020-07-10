package tokenize

import (
	"fmt"
	"strings"
)

//BaseTokenStream token stream
type BaseTokenStream struct {
	Tokens []BaseToken
	Offset int
	Level  int
}

//Tokenize tokenize a string
func (stream *BaseTokenStream) Tokenize(content string) {

	stream.Offset = 0

	runes := []rune(content)

	for _, rune := range runes {

		token := BaseToken{Content: string(rune)}

		stream.AddToken(token)
	}
}

//AddToken add token to stream
func (stream *BaseTokenStream) AddToken(token BaseToken) {

	stream.Tokens = append(stream.Tokens, token)
}

//AddTokenFromString split string to character and add each character as a token with type is providing type.
func (stream *BaseTokenStream) AddTokenFromString(tokenType int, str string) {

	for _, r := range []rune(str) {

		stream.AddToken(BaseToken{Type: tokenType, Content: string(r)})
	}
}

//AddTokenByContent add token
func (stream *BaseTokenStream) AddTokenByContent(content []rune, tokenType int) {

	stream.Tokens = append(stream.Tokens, BaseToken{Content: string(content), Type: tokenType})
}

//Debug print debug tree
func (stream *BaseTokenStream) Debug(level int, fnName func(int) string) {

	lastOffset := stream.Offset
	stream.ResetToBegin()

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
	stream.Offset = lastOffset
}

//DebugMark debug mark
func (stream *BaseTokenStream) DebugMark(level int, mark *Mark, ignores *[]int, fnName func(int) string) {

	length := mark.End - mark.Begin

	iterator := 0

	for {
		if length <= 0 || stream.EOS() {
			break
		}

		token := stream.GetTokenAt(mark.Begin + iterator)
		fmt.Printf("%s", ColorOffset(mark.Begin+iterator))
		if token != nil {

			for i := 0; i <= level; i++ {

				if i == 0 {

					fmt.Printf("|%s ", ColorType(token.Type))

				} else {

					fmt.Print("| ")
				}
			}

			if !isIgnoreInMark(mark.Begin+iterator, ignores) {

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

		iterator++
	}
}

//GetToken read token but not move pointer
func (stream *BaseTokenStream) GetToken() *BaseToken {

	if stream.Offset <= len(stream.Tokens)-1 {

		off := stream.Offset

		return &stream.Tokens[off]
	}
	return nil
}

//GetTokenIter get token at (offset + iterator) position
func (stream *BaseTokenStream) GetTokenIter(iterator int) *BaseToken {

	if stream.Offset+iterator <= len(stream.Tokens)-1 {

		off := stream.Offset + iterator

		return &stream.Tokens[off]
	}
	return nil
}

//GetTokenAt get token at offset
func (stream *BaseTokenStream) GetTokenAt(offset int) *BaseToken {

	if offset <= len(stream.Tokens)-1 {

		return &stream.Tokens[offset]
	}
	return nil
}

//FindPattern search pattern
func (stream *BaseTokenStream) FindPattern(patterns []Pattern, stopWhenFound bool, phraseBreak int, isIgnore func(int) bool, fnName func(int) string) []Mark {

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

				mark := Mark{Type: pattern.Type, Begin: stream.Offset, End: stream.Offset + iter, Ignores: ignores, Children: children}

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

				childMark.Begin = stream.Offset + iter

				log.Append(fmt.Sprintf("\n\t[%s %s] %s %t", ColorType(patternToken.Type), ColorName(fnName(patternToken.Type)), ColorContent(patternToken.Content), patternToken.IsPhraseUntil))
			}
			var match bool = true

			var moveIter int = 0

			nextToken := stream.GetTokenIter(iter)

			if nextToken == nil {
				break
			}

			if nextToken.Type == phraseBreak || isIgnore(nextToken.Type) {

				if pattern.IsRemoveGlobalIgnore || patternToken.IsIgnoreInResult {

					ignores = append(ignores, stream.Offset+iter)
				}
				iter++

				log.Append(fmt.Sprintf("\n"))

				continue
			}
			if patternToken.Content != "" {

				var currToken = stream.GetTokenIter(iter)

				if currToken == nil || currToken.Content != patternToken.Content {

					match = false

					log.Append(fmt.Sprintf("=>[%s %s %s]", ColorFail(), ColorType(currToken.Type), ColorContent(currToken.Content)))
				}
				if patternToken.IsIgnoreInResult {

					ignores = append(ignores, stream.Offset+iter+moveIter)
				}

				childMark.Begin = stream.Offset + iter

				moveIter = 1

			} else if patternToken.Type > 0 {

				var currToken = stream.GetTokenIter(iter)

				if currToken == nil || (currToken.Type != phraseBreak && currToken.Type != patternToken.Type) {

					match = false

					log.Append(fmt.Sprintf("=>[%s %s %s]", ColorFail(), ColorType(currToken.Type), ColorContent(currToken.Content)))
				}

				if patternToken.IsIgnoreInResult {

					ignores = append(ignores, stream.Offset+iter+moveIter)
				}

				if currToken.Type == patternToken.Type {

					childMark.Begin = stream.Offset + iter
				}

				moveIter = 1

			} else if patternToken.IsPhraseUntil {

				isWordFound := false

				for {
					var currToken = stream.GetTokenIter(iter + moveIter)

					if currToken == nil {

						match = false

						log.Append(fmt.Sprintf("=>[%s]", ColorFail()))

						break
					}
					if isIgnore(currToken.Type) {

						if pattern.IsRemoveGlobalIgnore || patternToken.IsIgnoreInResult {

							ignores = append(ignores, stream.Offset+iter+moveIter)
						}
						moveIter++

						continue
					}
					if currToken.Type == phraseBreak && isWordFound {

						if pattern.IsRemoveGlobalIgnore || patternToken.IsIgnoreInResult {

							ignores = append(ignores, stream.Offset+iter+moveIter)
						}
						moveIter++

						break

					} else if currToken.Type != phraseBreak && len(currToken.Content) > 0 {

						isWordFound = true
					}

					if patternToken.IsIgnoreInResult {

						ignores = append(ignores, stream.Offset+iter+moveIter)
					}

					moveIter++
				}
			}
			if !match {

				break
			}

			iter += moveIter

			childMark.End = stream.Offset + iter

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
func (stream *BaseTokenStream) GetMaskedToken(mark *Mark, ignores *[]int) *BaseToken {

	if mark.IsTokenStream {

		token := BaseToken{Type: mark.Type}

		len := mark.End - mark.Begin

		iterator := 0

		for {
			if len <= 0 || stream.EOS() {

				break
			}
			nextToken := stream.GetTokenAt(mark.Begin + iterator)

			if !isIgnoreInMark(mark.Begin+iterator, ignores) {

				token.Children.AddToken(*nextToken)

			}
			len--

			iterator++
		}

		return &token

	} else {

		len := mark.End - mark.Begin

		iterator := 0

		for {
			if len <= 0 || stream.EOS() {

				break
			}
			nextToken := stream.GetTokenAt(mark.Begin + iterator)

			if !isIgnoreInMark(mark.Begin+iterator, ignores) {

				return nextToken

			}
			len--

			iterator++
		}
	}
	return nil
}

//ReadToken read token
func (stream *BaseTokenStream) ReadToken() *BaseToken {

	if stream.Offset <= len(stream.Tokens)-1 {

		off := stream.Offset

		stream.Offset++

		return &stream.Tokens[off]
	}
	return nil
}

//ResetToBegin reset to begin
func (stream *BaseTokenStream) ResetToBegin() {

	stream.Offset = 0

	for _, token := range stream.Tokens {

		token.Children.ResetToBegin()
	}
}

//EOS is end of stream
func (stream *BaseTokenStream) EOS() bool {

	return stream.Offset >= len(stream.Tokens)
}

//Length get len of stream
func (stream *BaseTokenStream) Length() int {

	return len(stream.Tokens)
}

//ConcatStringContent concat content of tokens
func (stream *BaseTokenStream) ConcatStringContent() string {

	lastOffset := stream.Offset
	stream.ResetToBegin()

	content := ""

	for {
		if stream.EOS() {

			break
		}
		token := stream.ReadToken()

		content += string(token.Content)
	}
	stream.Offset = lastOffset
	return content
}

//ToArray get array of tokens
func (stream *BaseTokenStream) ToArray() []BaseToken {

	var rs []BaseToken
	lastOffset := stream.Offset
	stream.ResetToBegin()

	for {
		if stream.EOS() {

			break
		}
		token := stream.ReadToken()

		rs = append(rs, *token)
	}
	stream.Offset = lastOffset
	return rs
}

//ReadFirstTokenType read first token of type
func (stream *BaseTokenStream) ReadFirstTokenType(tokenType int) *BaseToken {

	stream.ResetToBegin()

	for {
		if stream.EOS() {

			break
		}
		token := stream.ReadToken()

		if token.Type == tokenType {

			return token
		}
	}

	return nil
}

//ReadNextTokenType read from current position to next match of token type
func (stream *BaseTokenStream) ReadNextTokenType(tokenType int) *BaseToken {

	for {
		if stream.EOS() {

			break
		}
		token := stream.ReadToken()

		if token.Type == tokenType {

			return token
		}
	}
	return nil
}
