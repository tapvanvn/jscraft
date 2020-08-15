package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"newcontinent-team.com/jscraft/entity"
	"newcontinent-team.com/jscraft/tokenize"
	"newcontinent-team.com/jscraft/tokenize/js"
)

var (
	usage           string                = "Usage: jscraft <template_dir> <layout_dir> <work_dir>"
	steps           chan entity.BuildStep = make(chan entity.BuildStep, 0)
	done            chan int              = make(chan int, 0)
	numStep         int                   = 0
	workID          int                   = 0
	hasError        bool                  = false
	compileContext  entity.CompileContext
	requireProvider chan *entity.JSScopeFile = make(chan *entity.JSScopeFile)
	requireCheck    chan *entity.CheckReady  = make(chan *entity.CheckReady)
)

func main() {

	var templateDir string
	var layoutDir string
	var workDir string
	var isDebug bool

	flag.BoolVar(&isDebug, "d", false, "debug")
	//flag.StringVar(&layoutDir, "layout", "", "layout path")
	//flag.StringVar(&workDir, "work", "", "work path")

	flag.Parse()

	if isDebug {
		fmt.Println("build debug")
	} else {
		fmt.Println("build release")
	}

	numArg := flag.NArg()

	if numArg < 3 {
		flag.Usage()
		os.Exit(1)
	}

	templateDir = flag.Arg(0)
	layoutDir = flag.Arg(1)
	workDir = flag.Arg(2)

	fmt.Println("template:", templateDir)
	fmt.Println("layout:", layoutDir)
	fmt.Println("work:", workDir)

	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		// path/to/whatever does not exist
		fmt.Println("template directory is not existed")
		os.Exit(1)
	}

	if _, err := os.Stat(layoutDir); os.IsNotExist(err) {
		// path/to/whatever does not exist
		fmt.Println("layout directory is not existed")
		os.Exit(1)
	}

	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		// path/to/whatever does not exist
		fmt.Println("work directory is not existed")
		os.Exit(1)
	}

	//read layout
	if _, err := os.Stat(layoutDir + "/layout.json"); os.IsNotExist(err) {
		fmt.Println("layout.json not found")
		os.Exit(1)
	}

	layoutData, err := ioutil.ReadFile(layoutDir + "/layout.json")

	if err != nil {
		fmt.Println("read layout.json error")
		os.Exit(1)
	}

	compileContext.Init()
	compileContext.LayoutDir = layoutDir
	compileContext.WorkDir = workDir
	compileContext.TemplateDir = templateDir

	compileContext.RequireProvider = &requireProvider
	compileContext.IsDebug = isDebug

	go processRequire()

	go processRequireCheck()

	var layout entity.Layout
	parseErr := json.Unmarshal(layoutData, &layout)

	if parseErr != nil {
		fmt.Println("layout bad syntax:" + parseErr.Error())
		os.Exit(1)
	}

	numStep = len(layout.BuildSteps)
	for _, step := range layout.BuildSteps {
		go addStep(step)
	}

	numProcessor := runtime.NumCPU()

	numWorker := numProcessor

	if numStep < numProcessor {
		numWorker = numStep
	}

	for i := 0; i < numWorker; i++ {
		go work()
	}

	for {
		if value, ok := <-done; ok {
			numStep += value

			if numStep == 0 {
				break
			}
		} else if numStep <= 0 {
			break
		}
	}
}

func processRequire() {

	var require_id = 0

	for {
		if hasError {

			return
		}
		select {

		case jsScopeFile := <-requireProvider:

			addBegin()

			//var current_require_id = require_id

			require_id++

			var jsmeaning entity.JSMeaning

			var jsmeaningHightContext entity.JSMeaningHighContext

			var uriMeaning entity.URIMeaning

			//fmt.Printf("\nprocess require %d : \n\t %s \n", current_require_id, jsScopeFile.FilePath)

			jsScopeFile.State = entity.FileStateLoading

			ext := filepath.Ext(jsScopeFile.FilePath)

			data, err := ioutil.ReadFile(jsScopeFile.FilePath)

			fmt.Println("load:", jsScopeFile.FilePath)

			if err != nil {

				hasError = true

				fmt.Println(err.Error())

				jsScopeFile.State = entity.FileStateError

				fmt.Println("\tfail")
				return
			}

			if strings.ToLower(ext) != ".js" {
				//Todo: error channel here
				hasError = true
				jsScopeFile.State = entity.FileStateError

				fmt.Println("\tfail")
				return
			}

			jsmeaning.Init(string(data), &compileContext)

			rawStream := tokenize.TokenStream{}

			for {
				token := jsmeaning.GetNextMeaningToken()

				if token == nil {
					break
				}
				rawStream.AddToken(*token)
			}

			//fmt.Printf("\ttoken:%d\n", rawStream.Length())

			jsmeaningHightContext.Init(rawStream, &compileContext)

			for {
				token := jsmeaningHightContext.GetNextMeaningToken()

				if token == nil {

					break
				}
				if token.Type == js.TokenJSCraft {

					jscraft := entity.GetJSCraft(token)

					if jscraft != nil {
						//fmt.Printf("begin jscraft %s \n", jscraft.FunctionName)
						if jscraft.FunctionName == "require" {

							requireURI := jscraft.Stream.ConcatStringContent()

							err := uriMeaning.Init(requireURI)
							if err != nil {
								//todo: process err
							}
							requirePath := compileContext.GetPathForNamespace(uriMeaning.Namespace) + "/" + uriMeaning.RelativePath

							//fmt.Printf("require %s\n", requirePath)

							scopeRequire := compileContext.RequireJSFile(requirePath)

							//fmt.Printf("add require to:%s \n", jsScopeFile.FilePath)

							checkReady := jsScopeFile.AddRequire(requirePath, scopeRequire)

							if checkReady != nil {

								go addCheckReady(checkReady)
							}

							//fmt.Println("after require")

						} else if jscraft.FunctionName == "template" {

							templateName := jscraft.GetTemplateName()

							templateToken := jscraft.GetTemplateToken()
							fmt.Printf("\nadd template 1: %s\n", templateName)
							jsScopeFile.AddTemplate(templateName, templateToken)
							fmt.Printf("\nfinish add template 1\n")
						}
						//fmt.Printf("end jscraft \n")
					}

				} else if token.Type == js.TokenJSFunction {

					jsfunc := entity.GetJSFunction(token)

					if jsfunc != nil {

						if len(jsfunc.FunctionName) > 8 && string(jsfunc.FunctionName[0:8]) == "jscraft_" {

							patchName := string(jsfunc.FunctionName[8:])

							patchStreamToken := tokenize.BaseToken{Type: js.TokenJSPatchStream, Children: jsfunc.Body.Children}

							compileContext.AddPatch(jsScopeFile.FilePath, patchName, patchStreamToken)
						}
					}
				}
				jsScopeFile.Stream.AddToken(*token)
			}

			//fmt.Printf("\tloaded : %d \n", current_require_id)

			jsScopeFile.State = entity.FileStateLoaded

			go addDone()
		}
	}
}

func processRequireCheck() {

	for {
		if hasError {

			return
		}
		select {

		case checkReady := <-requireCheck:

			addBegin()

			//fmt.Printf("check file: %s %d\n", checkReady.FileCheck.FilePath, checkReady.FileCheck.State)

			if checkReady.FileCheck.IsReady() {
				checkReady.IsReady = true
				//fmt.Printf("file ready: %s\n", checkReady.FileCheck.FilePath)
			} else {
				go addCheckReady(checkReady)
			}
			go addDone()
		}
	}
}

func addCheckReady(checkReady *entity.CheckReady) {
	requireCheck <- checkReady
}

func addStep(step entity.BuildStep) {

	steps <- step
}

func addBegin() {

	done <- 1
}
func addDone() {

	done <- -1
}

func work() {

	for {
		if hasError {

			return
		}

		select {

		case step := <-steps:
			var id = workID
			workID++

			fmt.Printf("\nbegin compile %d: from %s\n", id, step.From)

			var compiler entity.Compiler

			compiler.Init(step.Target, step.From, &compileContext)

			err := compiler.CompileTarget()

			fmt.Printf("\nfinish compile: %d\n\t to: %s\n", id, step.Target)

			if err != nil {

				fmt.Println(err.Error())
			}

			addDone()

		default:
			if numStep == 0 {

				return
			}
		}
	}
}
