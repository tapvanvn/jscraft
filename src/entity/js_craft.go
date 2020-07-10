package entity

import (
	"fmt"
	"log"

	"com.newcontinent-team.jscraft/tokenize"
	"com.newcontinent-team.jscraft/tokenize/js"
)

//JSCraft infomation about scraft call
type JSCraft struct {
	FunctionName string

	Stream *tokenize.BaseTokenStream
}

//GetTemplateName get template name of jscraft template function
func (jscraft *JSCraft) GetTemplateName() string {

	if jscraft.FunctionName == "template" {

		firstToken := jscraft.Stream.GetTokenAt(0)

		if firstToken != nil && firstToken.Type == js.TokenJSString {

			return firstToken.Children.ConcatStringContent()
		}
	}
	return ""
}

//GetTemplateToken get content of template
func (jscraft *JSCraft) GetTemplateToken() *tokenize.BaseToken {

	if jscraft.FunctionName == "template" {

		secondToken := jscraft.Stream.GetTokenAt(2)

		if secondToken.Type != js.TokenJSFunction && secondToken.Type != js.TokenJSFunctionLambda {

			return nil
		}

		jsfunc := GetJSFunction(secondToken)

		return &jsfunc.Body
	}
	return nil
}

//GetBuildTemplateName get template name of jscraft template function
func (jscraft *JSCraft) GetBuildTemplateName() string {

	if jscraft.FunctionName == "build" {

		firstToken := jscraft.Stream.GetTokenAt(0)

		if firstToken != nil && firstToken.Type == js.TokenJSString {

			return firstToken.Children.ConcatStringContent()
		}
	}
	return ""
}

//GetBuildBlockObject get template build parameter
func (jscraft *JSCraft) GetBuildBlockObject(patchContext *PatchContext) {

	if jscraft.FunctionName == "build" {

		secondToken := jscraft.Stream.GetTokenAt(2)

		if secondToken.Type == js.TokenJSBlock {

			secondToken.Children.ResetToBegin()
			fmt.Println("---------")
			secondToken.Children.Debug(0, js.TokenName)
			for {

				if secondToken.Children.EOS() {

					break
				}

				token := secondToken.Children.ReadToken()

				if token == nil {
					break
				}

				if token.Type != js.TokenJSString && token.Type != js.TokenJSWord {
					//error
					fmt.Println("continue:" + token.Content)
					continue
				}

				patchName := ""

				if token.Type == js.TokenJSString {

					patchName = token.Children.ConcatStringContent()

				} else {

					patchName = token.Content
				}
				if len(patchName) == 0 {
					continue
				}

				fmt.Println("pactName:" + patchName)

				_ = secondToken.Children.ReadToken()

				contentToken := secondToken.Children.ReadToken()

				fmt.Println(js.TokenName(contentToken.Type))

				contentBuildFunc := GetJSFunction(contentToken)

				if contentBuildFunc == nil {

					log.Fatal("Syntax Error 1" + contentToken.Content)
				}

				patchContext.AddPatch(patchName, contentBuildFunc.Body.Children)
			}
		}
	}

}
