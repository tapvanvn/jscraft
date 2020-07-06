package entity

import (
	"errors"
	"fmt"
	"time"

	"com.newcontinent-team.jscraft/tokenize"
	"com.newcontinent-team.jscraft/tokenize/js"
)

type JSBuildOptions struct {
	//for future use
	IsDebug bool
}

//JSBuilder builder
type JSBuilder struct {
	content string

	context *CompileContext

	options JSBuildOptions

	Error error

	cacheBuiltFile []string

	HighContextStream tokenize.BaseTokenStream

	fileScope *JSScopeFile
}

//Init init a build process
func (builder *JSBuilder) Init(fromFileScope *JSScopeFile, context *CompileContext, options JSBuildOptions) {

	builder.content = ""

	builder.context = context

	builder.options = options
	builder.options.IsDebug = true

	builder.fileScope = fromFileScope

	builder.process(fromFileScope)
}

//GetContent get build output
func (builder *JSBuilder) GetContent() string {

	return builder.content
}

func (builder *JSBuilder) process(fileScope *JSScopeFile) {

	if fileScope.State == FileStateLoading || fileScope.State == FileStateWaiting {

		for {

			time.Sleep(1 * time.Second)

			if fileScope.State != FileStateLoading && fileScope.State != FileStateWaiting {

				break
			} else {
				fmt.Printf("waiting for:%s curr:%d\n", fileScope.FilePath, fileScope.State)
			}
		}
	}

	if fileScope.State != FileStateLoaded {

		builder.Error = errors.New("file is not loaded:" + fileScope.FilePath)

		return
	}
	found := false

	stream := tokenize.BaseTokenStream{}

	for _, builtFile := range builder.cacheBuiltFile {

		if builtFile == fileScope.FilePath {

			found = true

			break
		}
	}

	if !found {

		builder.cacheBuiltFile = append(builder.cacheBuiltFile, fileScope.FilePath)

		fileScope.Stream.Debug(0, js.TokenName)
		builder.processStream(&fileScope.Stream, &stream)

		formatter := JSFormatter{}

		formatter.Format(&stream)

		builder.content += formatter.Content

		if builder.options.IsDebug {

			stream.Debug(0, js.TokenName)
		}
	}
}

func (builder *JSBuilder) processStream(stream *tokenize.BaseTokenStream, outStream *tokenize.BaseTokenStream) {

	stream.ResetToBegin()

	for {
		if stream.EOS() {

			break
		}
		token := stream.ReadToken()

		builder.processToken(token, outStream)
	}
}

func (builder *JSBuilder) processToken(token *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	switch token.Type {

	case js.TokenJSFunction, js.TokenJSFunctionLambda:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		builder.processFunction(token, outStream)

	case js.TokenJSFor:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		builder.processFor(token, outStream)

	case js.TokenJSWhile:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		builder.processWhile(token, outStream)

	case js.TokenJSDo:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		builder.processDo(token, outStream)

	case js.TokenJSSwitch:
		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		builder.processSwitch(token, outStream)

	case js.TokenJSIf:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		builder.processIf(token, outStream)

	case js.TokenJSElseIf:

		builder.processElseIf(token, outStream)

	case js.TokenJSElse:

		builder.processElse(token, outStream)

	case js.TokenJSBracketSquare:

		builder.processBracketSquare(token, outStream)

	case js.TokenJSBracket:

		builder.processBracket(token, outStream)

	case js.TokenJSBlock:

		builder.processBlock(token, outStream)

	case js.TokenJSCraft:

		builder.processCraft(token, outStream)

	case js.TokenJSPhrase:

		builder.processPhrase(token, outStream)

	case js.TokenJSString:

		builder.processString(token, outStream)

	case js.TokenJSRegex:

		builder.processRegex(token, outStream)

	case js.TokenJSWord:

		outStream.AddTokenFromString(js.TokenJSWord, token.Content)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSWordBreak})

	case js.TokenJSOperator, js.TokenJSUnaryOperator, js.TokenJSBinaryOperator, js.TokenJSAssign:

		outStream.AddTokenFromString(js.TokenJSOperator, token.Content)

	case js.TokenJSPhraseBreak:

		outStream.AddToken(*token)

	case js.TokenJSLineComment, js.TokenJSBlockComment:

		break

	case js.TokenJSCraftDebug:

		if builder.context.IsDebug {

			builder.processStream(&token.Children, outStream)
		}
	case js.TokenJSRightArrow:
		//todo: fix this later
		outStream.AddTokenFromString(js.TokenJSOperator, "=>")

	default:

		fmt.Printf("process token fail: %s %s %s\n", tokenize.ColorType(token.Type), tokenize.ColorName(js.TokenName(token.Type)), tokenize.ColorContent(token.Content))
		break
	}
}

func (builder *JSBuilder) processFunction(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	jsfunc := GetJSFunction(currToken)

	if jsfunc != nil {

		if len(jsfunc.FunctionName) == 0 {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

			builder.processBracket(&jsfunc.Params, outStream)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

			outStream.AddTokenFromString(js.TokenJSOperator, "=>")

			builder.processBlock(&jsfunc.Body, outStream)

		} else if len(jsfunc.FunctionName) <= 8 || string(jsfunc.FunctionName[0:8]) != "jscraft_" {

			outStream.AddTokenFromString(js.TokenJSWord, "function")

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSWordBreak})

			outStream.AddTokenFromString(js.TokenJSWord, jsfunc.FunctionName)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

			builder.processBracket(&jsfunc.Params, outStream)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

			builder.processBlock(&jsfunc.Body, outStream)
		}
		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

	} else {

		fmt.Printf("get JSFunction fail for token:%d %s\n", currToken.Type, currToken.Content)
	}
}

func (builder *JSBuilder) processFor(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	jsfor := GetJSFor(currToken)

	if jsfor != nil {

		outStream.AddTokenFromString(js.TokenJSWord, "for")

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

		builder.processBracket(&jsfor.Declare, outStream)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

		bodyToken := tokenize.BaseToken{}

		builder.processBlock(&jsfor.Body, &bodyToken.Children)

		outStream.AddToken(bodyToken)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

	} else {

		fmt.Printf("get JSFor fail for token:%d %s\n", currToken.Type, currToken.Content)
	}
}

func (builder *JSBuilder) processBlock(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	if currToken.Type == js.TokenJSBlock {

		outStream.AddTokenFromString(js.TokenJSOperator, "{")

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSScopeBegin})

		builder.processStream(&currToken.Children, outStream)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSScopeEnd})

		outStream.AddTokenFromString(js.TokenJSOperator, "}")
	}
}

func (builder *JSBuilder) processBracket(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	if currToken.Type == js.TokenJSBracket {

		outStream.AddTokenFromString(js.TokenJSOperator, "(")

		builder.processStream(&currToken.Children, outStream)

		outStream.AddTokenFromString(js.TokenJSOperator, ")")
	}
}

func (builder *JSBuilder) processBracketSquare(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	if currToken.Type == js.TokenJSBracketSquare {

		outStream.AddTokenFromString(js.TokenJSOperator, "[")

		builder.processStream(&currToken.Children, outStream)

		outStream.AddTokenFromString(js.TokenJSOperator, "]")
	}
}

func (builder *JSBuilder) processString(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	if currToken.Type == js.TokenJSString {

		outStream.AddTokenFromString(js.TokenJSWord, currToken.Content+currToken.Children.ConcatStringContent()+currToken.Content)
	}
}

func (builder *JSBuilder) processRegex(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	if currToken.Type == js.TokenJSRegex {

		outStream.AddTokenFromString(js.TokenJSWord, currToken.Children.ConcatStringContent())
	}
}

func (builder *JSBuilder) processIf(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	if currToken.Type == js.TokenJSIf {

		currToken.Children.ResetToBegin()

		firstToken := currToken.Children.GetToken()

		if firstToken == nil {
			//todo: error
			return
		}

		isNeedStrongBreak := firstToken.Type == js.TokenJSPhrase

		outStream.AddTokenFromString(js.TokenJSWord, "if")

		builder.processStream(&currToken.Children, outStream)

		if isNeedStrongBreak {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseStrongBreak})

		} else {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		}
	}
}

func (builder *JSBuilder) processElseIf(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	if currToken.Type == js.TokenJSElseIf {

		currToken.Children.ResetToBegin()

		firstToken := currToken.Children.GetToken()

		if firstToken == nil {
			//todo: error
			return
		}

		isNeedStrongBreak := firstToken.Type == js.TokenJSPhrase

		outStream.AddTokenFromString(js.TokenJSWord, "else if")

		builder.processStream(&currToken.Children, outStream)

		if isNeedStrongBreak {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseStrongBreak})

		} else {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		}
	}
}

func (builder *JSBuilder) processElse(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	if currToken.Type == js.TokenJSElse {

		currToken.Children.ResetToBegin()

		firstToken := currToken.Children.GetToken()

		if firstToken == nil {
			//todo: error
			return
		}

		isNeedStrongBreak := firstToken.Type == js.TokenJSPhrase

		outStream.AddTokenFromString(js.TokenJSWord, "else")

		builder.processStream(&currToken.Children, outStream)

		if isNeedStrongBreak {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseStrongBreak})

		} else {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		}
	}
}

func (builder *JSBuilder) processWhile(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	jswhile := GetJSWhile(currToken)

	if jswhile != nil {

		outStream.AddTokenFromString(js.TokenJSWord, "while")

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

		builder.processBracket(&jswhile.Condition, outStream)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

		if jswhile.Body.Type == js.TokenJSBlock {

			bodyToken := tokenize.BaseToken{}

			builder.processBlock(&jswhile.Body, &bodyToken.Children)

			outStream.AddToken(bodyToken)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

		} else {

			builder.processToken(&jswhile.Body, outStream)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseStrongBreak})
		}
	}
}

func (builder *JSBuilder) processDo(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	jsdo := GetJSDo(currToken)

	if jsdo != nil {

		outStream.AddTokenFromString(js.TokenJSWord, "do")

		if jsdo.Body.Type == js.TokenJSBlock {

			bodyToken := tokenize.BaseToken{}

			builder.processBlock(&jsdo.Body, &bodyToken.Children)

			outStream.AddToken(bodyToken)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

		} else {
			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSWordBreak})

			builder.processToken(&jsdo.Body, outStream)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseStrongBreak})
		}

		outStream.AddTokenFromString(js.TokenJSWord, "while")

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

		builder.processBracket(&jsdo.Condition, outStream)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
	}
}

func (builder *JSBuilder) processSwitch(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {
	jsswitch := GetJSSwitch(currToken)

	if jsswitch != nil {

		outStream.AddTokenFromString(js.TokenJSWord, "switch")

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

		builder.processBracket(&jsswitch.Var, outStream)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

		bodyToken := tokenize.BaseToken{}

		builder.processBlock(&jsswitch.Body, &bodyToken.Children)

		outStream.AddToken(bodyToken)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
	}
}

func (builder *JSBuilder) processPhrase(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	currToken.Children.ResetToBegin()

	for {
		if currToken.Children.EOS() {

			break
		}

		token := currToken.Children.ReadToken()

		builder.processToken(token, outStream)
	}

	outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
}

func (builder *JSBuilder) processCraft(currToken *tokenize.BaseToken, outStream *tokenize.BaseTokenStream) {

	jscraft := GetJSCraft(currToken)

	if jscraft != nil {

		switch jscraft.FunctionName {

		case "require":

			requireURI := jscraft.Stream.ConcatStringContent()

			path, err := builder.context.GetPathForURI(requireURI)

			if err != nil {

				return
			}
			scopeFile := builder.context.RequireJSFile(path)

			builder.process(scopeFile)

			break

		case "conflict":

			break

		case "fetch":

			name := jscraft.Stream.ConcatStringContent()

			patch := builder.context.GetPatch(builder.fileScope.FilePath, name)

			if patch != nil {

				builder.processStream(patch, outStream)

			} else {

				fmt.Println("patch not found:" + name)
			}
		}
	} else {
		fmt.Println("get craft fail")

	}
}
