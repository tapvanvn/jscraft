package entity

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//Compiler compier
type Compiler struct {
	Target string

	From string

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

	path, err := compiler.Context.GetPathForURI(compiler.Target)

	if err != nil {

		return err
	}
	compiler.Target = path

	path, err = compiler.Context.GetPathForURI(compiler.From)

	if err != nil {

		return err
	}
	compiler.From = path

	stat, err := os.Stat(compiler.From)

	if err != nil && os.IsNotExist(err) {

		return errors.New("src file not found")
	}
	isFromDir := stat.IsDir()

	ext := ""

	if !isFromDir {

		ext = filepath.Ext(compiler.From)
	}

	switch strings.ToLower(ext) {

	case ".js":

		jsScopeFile := compiler.Context.RequireJSFile(compiler.From)

		for {

			if jsScopeFile.IsReady() {

				var builder JSBuilder

				var buildOptions JSBuildOptions

				builder.Init(jsScopeFile, compiler.Context, buildOptions)

				if builder.Error == nil {

					err = ioutil.WriteFile(compiler.Target, []byte(builder.GetContent()), 0644)

					if err != nil {

						return err
					}
				} else {

					log.Println("some error:" + builder.Error.Error())
				}
				break
			}

			time.Sleep(time.Millisecond * 100)
		}

	default:
		if !isFromDir {

			data, err := ioutil.ReadFile(compiler.From)

			if err != nil {

				return err
			}
			err = ioutil.WriteFile(compiler.Target, data, 0644)

			if err != nil {

				return err
			}
		} else {

			cmd := exec.Command("cp", "-Rf", compiler.From, compiler.Target)

			err := cmd.Run()

			if err != nil {

				return err
			}
		}
	}

	return nil
}
