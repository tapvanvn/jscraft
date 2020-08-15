package entity

import (
	"log"
	"strconv"

	"newcontinent-team.com/jscraft/tokenize"
	"newcontinent-team.com/jscraft/tokenize/js"
)

//JSCraft infomation about scraft call
type JSCraft struct {
	FunctionName string

	Stream *tokenize.TokenStream
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

			iterator := secondToken.Children.Iterator()

			for {

				if iterator.EOS() {

					break
				}

				token := iterator.ReadToken()

				if token == nil {
					break
				}

				if token.Type != js.TokenJSString && token.Type != js.TokenJSWord {

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

				_ = iterator.ReadToken()

				contentToken := iterator.ReadToken()

				if contentToken == nil {

					log.Fatalf("Syntax Error \n")
				}

				//fmt.Println(js.TokenName(contentToken.Type))

				if contentToken.Type == js.TokenJSFunction || contentToken.Type == js.TokenJSFunctionLambda {

					contentBuildFunc := GetJSFunction(contentToken)

					if contentBuildFunc == nil {

						log.Fatalf("Syntax Error : %d \n", strconv.Itoa(contentToken.Type))
					}

					patchStreamToken := tokenize.BaseToken{Type: js.TokenJSPatchStream, Children: contentBuildFunc.Body.Children}

					patchContext.AddPatch(patchName, patchStreamToken)

				} else if contentToken.Type == js.TokenJSString || contentToken.Type == js.TokenJSWord {

					patchContext.AddPatch(patchName, *contentToken)
				}

			}
		}
	}

}
