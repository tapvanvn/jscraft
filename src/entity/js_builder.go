package entity

import (
	"errors"
	"fmt"
	"log"
	"time"

	"newcontinent-team.com/jscraft/tokenize"
	"newcontinent-team.com/jscraft/tokenize/js"
)

//JSBuildOptions build option
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

	HighContextStream tokenize.TokenStream

	fileScope *JSScopeFile

	builderContext *BuilderContext

	patchContext *PatchContext
}

//Init init a build process
func (builder *JSBuilder) Init(fromFileScope *JSScopeFile, context *CompileContext, options JSBuildOptions) {

	builder.builderContext = context.MakeBuildContext(fromFileScope)

	builder.patchContext = context.MakePatchContext(fromFileScope)

	if builder.builderContext == nil {

		log.Fatal("create build context fail: " + fromFileScope.FilePath)
	}

	builder.content = ""

	builder.context = context

	builder.options = options

	builder.options.IsDebug = true

	builder.fileScope = fromFileScope

	builder.process(fromFileScope, builder.patchContext)

}

//GetContent get build output
func (builder *JSBuilder) GetContent() string {

	return builder.content
}

func (builder *JSBuilder) process(fileScope *JSScopeFile, patchContext *PatchContext) {

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

	fmt.Printf(" processing file %s\n\n", fileScope.FilePath)

	//debug.PrintStack()

	found := false

	stream := tokenize.TokenStream{}

	for _, builtFile := range builder.cacheBuiltFile {

		if builtFile == fileScope.FilePath {

			found = true

			break
		}
	}

	if !found {

		builder.cacheBuiltFile = append(builder.cacheBuiltFile, fileScope.FilePath)

		fmt.Printf(" build file %s\n\n", fileScope.FilePath)

		//fileScope.Stream.Debug(0, js.TokenName)

		iterator := fileScope.Stream.Iterator()

		builder.processStream(&iterator, &stream, patchContext)

		formatter := JSFormatter{}

		formatter.Format(&stream)

		builder.content += formatter.Content

		if builder.options.IsDebug {

			//stream.Debug(0, js.TokenName)
		}
	}
}

func (builder *JSBuilder) processStream(iterator *tokenize.TokenStreamIterator, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	iterator.ResetToBegin()

	for {
		if iterator.EOS() {

			break
		}
		token := iterator.ReadToken()

		builder.processToken(token, outStream, patchContext)
	}
}

func (builder *JSBuilder) processToken(token *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	switch token.Type {

	case js.TokenJSFunction, js.TokenJSFunctionLambda:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

		builder.processFunction(token, outStream, patchContext)

	case js.TokenJSFor:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

		builder.processFor(token, outStream, patchContext)

	case js.TokenJSWhile:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

		builder.processWhile(token, outStream, patchContext)

	case js.TokenJSDo:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

		builder.processDo(token, outStream, patchContext)

	case js.TokenJSSwitch:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

		builder.processSwitch(token, outStream, patchContext)

	case js.TokenJSIf:

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

		builder.processIf(token, outStream, patchContext)

	case js.TokenJSElseIf:

		builder.processElseIf(token, outStream, patchContext)

	case js.TokenJSElse:

		builder.processElse(token, outStream, patchContext)

	case js.TokenJSBracketSquare:

		builder.processBracketSquare(token, outStream, patchContext)

	case js.TokenJSBracket:

		builder.processBracket(token, outStream, patchContext)

	case js.TokenJSBlock:

		builder.processBlock(token, outStream, patchContext)

	case js.TokenJSCraft:

		builder.processCraft(token, outStream, patchContext)

	case js.TokenJSPhrase:

		builder.processPhrase(token, outStream, patchContext)

	case js.TokenJSString:

		builder.processString(token, outStream, patchContext)

	case js.TokenJSRegex:

		builder.processRegex(token, outStream, patchContext)

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

			iterator := token.Children.Iterator()
			builder.processStream(&iterator, outStream, patchContext)
		}
	case js.TokenJSRightArrow:
		//todo: fix this later
		outStream.AddTokenFromString(js.TokenJSOperator, "=>")

	default:

		fmt.Printf("process token fail: %s %s %s\n", tokenize.ColorType(token.Type), tokenize.ColorName(js.TokenName(token.Type)), tokenize.ColorContent(token.Content))
		break
	}
}

func (builder *JSBuilder) processFunction(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	jsfunc := GetJSFunction(currToken)

	if jsfunc != nil {

		if len(jsfunc.FunctionName) == 0 {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

			builder.processBracket(&jsfunc.Params, outStream, patchContext)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

			outStream.AddTokenFromString(js.TokenJSOperator, "=>")

			builder.processBlock(&jsfunc.Body, outStream, patchContext)

		} else if len(jsfunc.FunctionName) <= 8 || string(jsfunc.FunctionName[0:8]) != "jscraft_" {

			outStream.AddTokenFromString(js.TokenJSWord, "function")

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSWordBreak})

			outStream.AddTokenFromString(js.TokenJSWord, jsfunc.FunctionName)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

			builder.processBracket(&jsfunc.Params, outStream, patchContext)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

			builder.processBlock(&jsfunc.Body, outStream, patchContext)
		}
		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

	} else {

		fmt.Printf("get JSFunction fail for token:%d %s\n", currToken.Type, currToken.Content)
	}
}

func (builder *JSBuilder) processFor(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	jsfor := GetJSFor(currToken)

	if jsfor != nil {

		outStream.AddTokenFromString(js.TokenJSWord, "for")

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

		builder.processBracket(&jsfor.Declare, outStream, patchContext)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

		bodyToken := tokenize.BaseToken{}

		builder.processBlock(&jsfor.Body, &bodyToken.Children, patchContext)

		outStream.AddToken(bodyToken)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

	} else {

		fmt.Printf("get JSFor fail for token:%d %s\n", currToken.Type, currToken.Content)
	}
}

func (builder *JSBuilder) processBlock(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	if currToken.Type == js.TokenJSBlock {

		outStream.AddTokenFromString(js.TokenJSOperator, "{")

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSScopeBegin})

		iterator := currToken.Children.Iterator()

		builder.processStream(&iterator, outStream, patchContext)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSScopeEnd})

		outStream.AddTokenFromString(js.TokenJSOperator, "}")
	}
}

func (builder *JSBuilder) processBracket(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	if currToken.Type == js.TokenJSBracket {

		outStream.AddTokenFromString(js.TokenJSOperator, "(")

		iterator := currToken.Children.Iterator()

		builder.processStream(&iterator, outStream, patchContext)

		outStream.AddTokenFromString(js.TokenJSOperator, ")")
	}
}

func (builder *JSBuilder) processBracketSquare(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	if currToken.Type == js.TokenJSBracketSquare {

		outStream.AddTokenFromString(js.TokenJSOperator, "[")

		iterator := currToken.Children.Iterator()

		builder.processStream(&iterator, outStream, patchContext)

		outStream.AddTokenFromString(js.TokenJSOperator, "]")
	}
}

func (builder *JSBuilder) processString(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	if currToken.Type == js.TokenJSString {

		outStream.AddTokenFromString(js.TokenJSWord, currToken.Content+currToken.Children.ConcatStringContent()+currToken.Content)
	}
}

func (builder *JSBuilder) processRegex(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	if currToken.Type == js.TokenJSRegex {

		outStream.AddTokenFromString(js.TokenJSWord, currToken.Children.ConcatStringContent())
	}
}

func (builder *JSBuilder) processIf(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	if currToken.Type == js.TokenJSIf {

		iterator := currToken.Children.Iterator()

		firstToken := iterator.GetToken()

		if firstToken == nil {
			//todo: error
			return
		}

		isNeedStrongBreak := firstToken.Type == js.TokenJSPhrase

		outStream.AddTokenFromString(js.TokenJSWord, "if")

		iterator.ResetToBegin()

		builder.processStream(&iterator, outStream, patchContext)

		if isNeedStrongBreak {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseStrongBreak})

		} else {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		}
	}
}

func (builder *JSBuilder) processElseIf(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	if currToken.Type == js.TokenJSElseIf {

		iterator := currToken.Children.Iterator()

		firstToken := iterator.GetToken()

		if firstToken == nil {
			//todo: error
			return
		}

		isNeedStrongBreak := firstToken.Type == js.TokenJSPhrase

		outStream.AddTokenFromString(js.TokenJSWord, "else if")

		iterator.ResetToBegin()

		builder.processStream(&iterator, outStream, patchContext)

		if isNeedStrongBreak {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseStrongBreak})

		} else {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		}
	}
}

func (builder *JSBuilder) processElse(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	if currToken.Type == js.TokenJSElse {

		iterator := currToken.Children.Iterator()

		firstToken := iterator.GetToken()

		if firstToken == nil {
			//todo: error
			return
		}

		isNeedStrongBreak := firstToken.Type == js.TokenJSPhrase

		outStream.AddTokenFromString(js.TokenJSWord, "else")

		iterator.ResetToBegin()

		builder.processStream(&iterator, outStream, patchContext)

		if isNeedStrongBreak {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseStrongBreak})

		} else {

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
		}
	}
}

func (builder *JSBuilder) processWhile(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	jswhile := GetJSWhile(currToken)

	if jswhile != nil {

		outStream.AddTokenFromString(js.TokenJSWord, "while")

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

		builder.processBracket(&jswhile.Condition, outStream, patchContext)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

		if jswhile.Body.Type == js.TokenJSBlock {

			bodyToken := tokenize.BaseToken{}

			builder.processBlock(&jswhile.Body, &bodyToken.Children, patchContext)

			outStream.AddToken(bodyToken)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

		} else {

			builder.processToken(&jswhile.Body, outStream, patchContext)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseStrongBreak})
		}
	}
}

func (builder *JSBuilder) processDo(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	jsdo := GetJSDo(currToken)

	if jsdo != nil {

		outStream.AddTokenFromString(js.TokenJSWord, "do")

		if jsdo.Body.Type == js.TokenJSBlock {

			bodyToken := tokenize.BaseToken{}

			builder.processBlock(&jsdo.Body, &bodyToken.Children, patchContext)

			outStream.AddToken(bodyToken)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})

		} else {
			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSWordBreak})

			builder.processToken(&jsdo.Body, outStream, patchContext)

			outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseStrongBreak})
		}

		outStream.AddTokenFromString(js.TokenJSWord, "while")

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

		builder.processBracket(&jsdo.Condition, outStream, patchContext)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
	}
}

func (builder *JSBuilder) processSwitch(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	jsswitch := GetJSSwitch(currToken)

	if jsswitch != nil {

		outStream.AddTokenFromString(js.TokenJSWord, "switch")

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueBegin})

		builder.processBracket(&jsswitch.Var, outStream, patchContext)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSGlueEnd})

		bodyToken := tokenize.BaseToken{}

		builder.processBlock(&jsswitch.Body, &bodyToken.Children, patchContext)

		outStream.AddToken(bodyToken)

		outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
	}
}

func (builder *JSBuilder) processPhrase(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	iterator := currToken.Children.Iterator()

	for {

		if iterator.EOS() {

			break
		}

		token := iterator.ReadToken()

		builder.processToken(token, outStream, patchContext)
	}

	outStream.AddToken(tokenize.BaseToken{Type: js.TokenJSPhraseBreak})
}

func (builder *JSBuilder) processCraft(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	jscraft := GetJSCraft(currToken)

	if jscraft != nil {

		switch jscraft.FunctionName {

		case "require":

			requireURI := jscraft.Stream.ConcatStringContent()

			path, err := builder.context.GetPathForURI(requireURI)

			if err != nil {
				fmt.Println("jscraft require error" + err.Error())
				return
			}

			scopeFile := builder.context.RequireJSFile(path)

			//fmt.Printf("require file: %s \n\t==>%s\n", builder.fileScope.FilePath, scopeFile.FilePath)

			builder.process(scopeFile, patchContext)

			break

		case "conflict":

			break

		case "fetch":

			name := jscraft.Stream.ConcatStringContent()

			patch := patchContext.GetPatch(name)

			if patch != nil {

				if patchContext.Parent != nil {

					builder.processPatch(patch, outStream, patchContext.Parent)

				} else {

					builder.processPatch(patch, outStream, patchContext)
				}

			} else {

				fmt.Println("patch not found:" + name)
			}

		case "build":

			templateName := jscraft.GetBuildTemplateName()

			templateToken := builder.builderContext.GetTemplate(templateName)

			if templateToken == nil {

				//builder.builderContext.Debug()
				//builder.builderContext.Context.DebugDependence(builder.builderContext.FileScope)
				//builder.builderContext.FileScope.Debug()
				//patchContext.Debug()
				//patchContext.Context.Debug()
				log.Fatal("syntax error :" + templateName)

			}

			buildPatchContext := &PatchContext{}

			buildPatchContext.Init(patchContext, builder.context)

			jscraft.GetBuildBlockObject(buildPatchContext)

			iterator := templateToken.Children.Iterator()

			builder.processStream(&iterator, outStream, buildPatchContext)

			//templateToken.Children.Debug(0, js.TokenName)
			break
		}
	} else {

		fmt.Println("get craft fail")

	}
}

func (builder *JSBuilder) processPatch(currToken *tokenize.BaseToken, outStream *tokenize.TokenStream, patchContext *PatchContext) {

	if currToken.Type == js.TokenJSPatchStream {

		iterator := currToken.Children.Iterator()

		builder.processStream(&iterator, outStream, patchContext)

	} else {

		builder.processToken(currToken, outStream, patchContext)
	}
}
