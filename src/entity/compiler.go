package entity

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//Compiler compier
type Compiler struct {
	Target  string
	From    string
	Context *CompileContext
}

//Init init
func (compiler *Compiler) Init(target string, from string, context *CompileContext) {
	compiler.Target = target
	compiler.From = from
	compiler.Context = context
}

//CompileTarget parse target
func (compiler *Compiler) CompileTarget() error {
	var meaning URIMeaning
	err := meaning.Init(compiler.Target)
	if err != nil {
		return err
	}
	compiler.Target = compiler.Context.GetPathForNamespace(meaning.Namespace) + "/" + meaning.RelativePath

	//fmt.Println("target:" + compiler.Target)
	err = meaning.Init(compiler.From)
	if err != nil {
		return err
	}
	compiler.From = compiler.Context.GetPathForNamespace(meaning.Namespace) + "/" + meaning.RelativePath
	//fmt.Println("from:" + compiler.From)

	stat, err := os.Stat(compiler.From)
	if err != nil && os.IsNotExist(err) {
		return errors.New("src file not found")
	}

	if stat.IsDir() {
		return errors.New("src file is directory")
	}

	ext := filepath.Ext(compiler.From)

	switch strings.ToLower(ext) {
	case ".js":
		//fmt.Println("compile js")
		_ = compiler.Context.RequireJSFile(compiler.From)
		/*var jsmeaning JSMeaning
		jsmeaning.Init(string(data), compiler.Context)
		for {
			token := jsmeaning.GetNextMeaningToken()
			if token == nil {
				break
			}
			fmt.Println(token.Content)
			//jsStream.AddToken(token)
		}*/
	default:
		data, err := ioutil.ReadFile(compiler.From)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(compiler.Target, data, 0644)
		if err != nil {
			return err
		}
	}

	/*
		meaning.Stream.ResetToBegin()
		jsStream tokenize.BaseTokenStream

		for {
			if meaning.Stream.EOS() {
				break
			}
			token := meaning.Stream.GetNextMeaningToken()
			jsStream.AddToken(token)
		}
	*/

	//fmt.Println("from:" + ext)
	return nil
}
