package tokenize

import (
	"fmt"
	"strings"
)

//BaseTokenStream token stream
type BaseTokenStream struct {
	Tokens []BaseToken
	Offset int
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
	//fmt.Println("tokendfsdfk:" + strconv.Itoa(len(stream.Tokens)))
}

//AddTokenByContent add token
func (stream *BaseTokenStream) AddTokenByContent(content []rune, tokenType int) {

	stream.Tokens = append(stream.Tokens, BaseToken{Content: string(content), Type: tokenType})
	//fmt.Println("tokendfsdfk:" + strconv.Itoa(len(stream.Tokens)))
}

//Debug Debug
func (stream *BaseTokenStream) Debug(level int) {
	stream.ResetToBegin()
	for _, token := range stream.Tokens {
		trimContent := strings.Trim(token.Content, " \n\r")
		if len(trimContent) > 0 || token.Children.Length() > 0 {
			for i := 0; i <= level; i++ {
				if i == 0 {
					fmt.Printf("|%2d ", token.Type)
				} else {
					fmt.Print("| ")
				}
			}
			if len(trimContent) > 0 {
				fmt.Println(token.Content)
			} else {
				fmt.Println("")
			}

		}

		token.Children.Debug(level + 1)
	}
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
