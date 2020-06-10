package entity

import (
	"errors"
	"fmt"

	"com.newcontinent-team.jscraft/tokenize"
	"com.newcontinent-team.jscraft/tokenize/js"
)

type JSBuildOptions struct {
	//for future use
}

//JSBuilder builder
type JSBuilder struct {
	content        string
	context        *CompileContext
	options        JSBuildOptions
	Error          error
	cacheBuiltFile []string
}

//Init init a build process
func (builder *JSBuilder) Init(fromFileScope *JSScopeFile, context *CompileContext, options JSBuildOptions) {
	builder.content = ""
	builder.context = context
	builder.options = options

	builder.process(fromFileScope)
}

//GetContent get built output
func (builder *JSBuilder) GetContent() string {
	return builder.content
}

func (builder *JSBuilder) process(fileScope *JSScopeFile) {
	if !fileScope.IsLoaded {
		builder.Error = errors.New("file is not loaded:" + fileScope.FilePath)
		return
	}
	found := false
	for _, builtFile := range builder.cacheBuiltFile {
		if builtFile == fileScope.FilePath {
			found = true
			break
		}
	}
	if !found {

		builder.cacheBuiltFile = append(builder.cacheBuiltFile, fileScope.FilePath)
		builder.content += builder.processStream(&fileScope.Stream)

	}
}

//beautyAppend add a tab at begin of line of content
func (builder *JSBuilder) beautyAppend(content string) string {
	return content
}

func (builder *JSBuilder) processStream(stream *tokenize.BaseTokenStream) string {
	content := ""
	stream.ResetToBegin()
	for {
		if stream.EOS() {
			break
		}
		token := stream.ReadToken()
		switch token.Type {
		case js.TokenJSFunction:
			content += builder.processFunction(token)
		case js.TokenJSBlock:
			content += builder.processBlock(token)
		case js.TokenJSBracket:
			content += builder.processBracket(token)
		case js.TokenJSBracketSquare:
			content += builder.processBracketSquare(token)
		case js.TokenJSFor:
			content += builder.processFor(token)
		case js.TokenJSIf:
			content += builder.processIf(token)
		case js.TokenJSCraft:
			content += builder.processCraft(token)
		case js.TokenJSVariable:
			content += builder.processVariable(token)
		case js.TokenJSString:
			content += builder.processString(token)
		case js.TokenJSCraftDebug:
			if builder.context.IsDebug {
				content += builder.processStream(&token.Children)
			}
		default:
			content += token.Content
		}
	}
	return content
}

//todo: should we check error here?
func (builder *JSBuilder) processFunction(currToken *tokenize.BaseToken) string {
	content := ""
	if currToken.Type == js.TokenJSFunction {
		funcName := currToken.Content
		if len(funcName) > 8 && string(funcName[0:8]) == "jscraft_" {

		} else {
			content += "function " + currToken.Content
			children := currToken.Children.ToArray()
			for _, token := range children {
				if token.Type == js.TokenJSFunctionParam {
					content += builder.processFunctionParam(&token)
				} else if token.Type == js.TokenJSBlock {
					content += builder.processBlock(&token)
				}
			}
			content += "\n"
		}
	}
	return content
}

func (builder *JSBuilder) processBlock(currToken *tokenize.BaseToken) string {
	content := ""
	if currToken.Type == js.TokenJSBlock {
		content += "{\n"
		content += builder.beautyAppend(builder.processStream(&currToken.Children))
		content += "}"
	}
	return content
}

func (builder *JSBuilder) processFunctionParam(currToken *tokenize.BaseToken) string {
	content := ""
	if currToken.Type == js.TokenJSFunctionParam {
		params := currToken.Children.ToArray()
		content += "("
		numParams := len(params)
		if numParams > 0 {
			for i, paramToken := range params {
				if i > 0 {
					content += ","
				}
				content += paramToken.Content
			}
		}
		content += ")"
	}
	return content
}

func (builder *JSBuilder) processBracket(currToken *tokenize.BaseToken) string {
	content := ""
	if currToken.Type == js.TokenJSBracket {
		content += "("
		content += builder.processStream(&currToken.Children)
		content += ")"
	}
	return content
}

func (builder *JSBuilder) processBracketSquare(currToken *tokenize.BaseToken) string {
	content := ""
	if currToken.Type == js.TokenJSBracketSquare {
		content += "["
		content += builder.processStream(&currToken.Children)
		content += "]"
	}
	return content
}

func (builder *JSBuilder) processString(currToken *tokenize.BaseToken) string {
	content := ""

	if currToken.Type == js.TokenJSString {
		content += currToken.Content + currToken.Children.ConcatStringContent() + currToken.Content
	}
	return content
}

func (builder *JSBuilder) processFor(currToken *tokenize.BaseToken) string {
	content := ""
	if currToken.Type == js.TokenJSFor {
		content += "for"
		forChildren := currToken.Children.ToArray()

		for _, forChildToken := range forChildren {

			if forChildToken.Type == js.TokenJSBracket {

				content += builder.processBracket(&forChildToken)

			} else if forChildToken.Type == js.TokenJSBlock {

				content += builder.processBlock(&forChildToken)
			}
		}
	}
	return content
}

func (builder *JSBuilder) processIf(currToken *tokenize.BaseToken) string {
	content := ""
	if currToken.Type == js.TokenJSIf {
		content += "if"
		content += builder.processStream(&currToken.Children)
	}
	return content
}

func GetRequireURI(token *tokenize.BaseToken) (string, error) {
	if token.Type == js.TokenJSCraft && token.Content == "require" {
		bracketToken := token.Children.ReadFirstTokenType(js.TokenJSBracket)
		if bracketToken == nil {
			return "", errors.New("invalid token")
		}
		stringToken := bracketToken.Children.ReadFirstTokenType(js.TokenJSString)
		if stringToken == nil {
			return "", errors.New("invalid token")
		}
		return stringToken.Children.ConcatStringContent(), nil
	}
	return "", errors.New("token is invalid")
}

//GetPatchNameOfFetchCommand ...
func GetPatchNameOfFetchCommand(token *tokenize.BaseToken) (string, error) {
	if token.Type == js.TokenJSCraft && token.Content == "fetch" {
		bracketToken := token.Children.ReadFirstTokenType(js.TokenJSBracket)
		if bracketToken == nil {
			return "", errors.New("invalid token")
		}
		stringToken := bracketToken.Children.ReadFirstTokenType(js.TokenJSString)
		if stringToken == nil {
			return "", errors.New("invalid token")
		}
		return stringToken.Children.ConcatStringContent(), nil
	}
	return "", errors.New("token is invalid")
}

func (builder *JSBuilder) processCraft(currToken *tokenize.BaseToken) string {
	content := ""
	if currToken.Type == js.TokenJSCraft {
		switch currToken.Content {
		case "require":
			requireURI, err := GetRequireURI(currToken)
			if err != nil {
				//error
				return ""
			}
			path, err := builder.context.GetPathForURI(requireURI)
			if err != nil {
				//error
				return ""
			}
			scopeFile := builder.context.RequireJSFile(path)
			builder.process(scopeFile)
			break
		case "conflict":
			break
		case "fetch":
			name, err := GetPatchNameOfFetchCommand(currToken)
			if err != nil {
				fmt.Println(err.Error())
				return ""
			}
			patch := builder.context.GetPatch(name)
			if patch != nil {
				content += builder.processStream(patch)
			} else {
				fmt.Println("patch not found:" + name)
			}
		}
	}
	return content
}

func (builder *JSBuilder) processVariable(currToken *tokenize.BaseToken) string {
	content := ""
	if currToken.Type == js.TokenJSVariable {
		content += "var" + builder.processStream(&currToken.Children)
	}
	return content
}
