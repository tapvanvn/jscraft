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

	"com.newcontinent-team.jscraft/entity"
	"com.newcontinent-team.jscraft/tokenize"
	"com.newcontinent-team.jscraft/tokenize/js"
)

var (
	usage           string                = "Usage: jscraft <template_dir> <layout_dir> <work_dir>"
	steps           chan entity.BuildStep = make(chan entity.BuildStep, 0)
	done            chan int              = make(chan int, 0)
	numStep         int                   = 0
	workID          int                   = 0
	hasError        bool                  = false
	compileContext  entity.CompileContext
	requireProvider chan *entity.JSScopeFile = make(chan *entity.JSScopeFile, 0)
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
	compileContext.CacheProvider = make(map[string]*entity.JSScopeFile)
	compileContext.RequireProvider = &requireProvider
	compileContext.IsDebug = isDebug

	go processRequire()

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
			//fmt.Println("remainStep:" + strconv.Itoa(numStep))
			if numStep == 0 {
				break
			}
		} else if numStep <= 0 {
			break
		}

	}
}

func processRequire() {

	var jsmeaning entity.JSMeaning

	var jsmeaningHightContext entity.JSMeaningHighContext

	var uriMeaning entity.URIMeaning

	for {
		if hasError {

			return
		}
		select {

		case jsScopeFile := <-requireProvider:

			addBegin()

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

			rawStream := tokenize.BaseTokenStream{}

			for {
				token := jsmeaning.GetNextMeaningToken()

				if token == nil {
					break
				}
				rawStream.AddToken(*token)
			}

			fmt.Printf("\ttoken:%d\n", rawStream.Length())

			jsmeaningHightContext.Init(rawStream, &compileContext)

			for {
				token := jsmeaningHightContext.GetNextMeaningToken()

				if token == nil {

					break
				}
				if token.Type == js.TokenJSCraft {

					jscraft := entity.GetJSCraft(token)

					if jscraft != nil {

						if jscraft.FunctionName == "require" {

							requireURI := jscraft.Stream.ConcatStringContent()

							err := uriMeaning.Init(requireURI)
							if err != nil {
								//todo: process err
							}
							requirePath := compileContext.GetPathForNamespace(uriMeaning.Namespace) + "/" + uriMeaning.RelativePath

							jsScopeFile.Requires[requirePath] = compileContext.RequireJSFile(requirePath)

						}
					}
				} else if token.Type == js.TokenJSFunction {

					jsfunc := entity.GetJSFunction(token)

					if jsfunc != nil {

						if len(jsfunc.FunctionName) > 8 && string(jsfunc.FunctionName[0:8]) == "jscraft_" {

							patchName := string(jsfunc.FunctionName[8:])
							fmt.Println("patch:", patchName)

							compileContext.AddPatch(jsScopeFile.FilePath, patchName, jsfunc.Body.Children)

						}
					}
				}
				jsScopeFile.Stream.AddToken(*token)
			}

			fmt.Println("\tloaded")

			jsScopeFile.State = entity.FileStateLoaded

			go addDone()
		}
	}
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
	//var id = workID
	//workID++
	var compiler entity.Compiler

	for {
		if hasError {

			return
		}

		select {

		case step := <-steps:
			//fmt.Println(strconv.Itoa(id) + ":from:" + step.From)
			compiler.Init(step.Target, step.From, &compileContext)

			err := compiler.CompileTarget()

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
