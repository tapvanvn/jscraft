package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"runtime"

	"com.newcontinent-team.jscraft/entity"
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

	numCpu := runtime.NumCPU()

	numWorker := numCpu

	if numStep < numCpu {
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

	for {
		if hasError {
			return
		}
		select {
		case jsScopeFile := <-requireProvider:

			//fmt.Println("process require:" + jsScopeFile.FilePath)

			addBegin()

			ext := filepath.Ext(jsScopeFile.FilePath)
			data, err := ioutil.ReadFile(jsScopeFile.FilePath)

			if err != nil {
				hasError = true
				fmt.Println(err.Error())
				return
			}

			if strings.ToLower(ext) != ".js" {
				//Todo: error channel here
				hasError = true
				return
			}

			var jsmeaning entity.JSMeaning
			jsmeaning.Init(string(data), &compileContext)

			fmt.Println("--------")

			for {
				token := jsmeaning.GetNextMeaningToken()
				if token == nil {
					break
				}

				if token.Type == js.TokenJSCraft {
					if token.Content == "require" {
						blockToken := token.Children.ReadToken()
						if blockToken.Type == js.TokenJSBracket {
							stringToken := blockToken.Children.ReadToken()
							if stringToken.Type == js.TokenJSString {
								requireURI := stringToken.Children.ConcatStringContent()
								fmt.Println("require:" + requireURI)
								var uriMeaning entity.URIMeaning
								err := uriMeaning.Init(requireURI)
								if err != nil {
									//todo: process err
								}
								requirePath := compileContext.GetPathForNamespace(uriMeaning.Namespace) + "/" + uriMeaning.RelativePath
								jsScopeFile.Requires[requirePath] = compileContext.RequireJSFile(requirePath)
							}
						}
					} else if token.Content == "fetch" {

					}
				} else if token.Type == js.TokenJSFunction {
					funcName := token.Content
					//fmt.Println("func:" + funcName)
					if len(funcName) > 8 && string(funcName[0:8]) == "jscraft_" {
						patchName := string(funcName[8:])

						funcTokens := token.Children.ToArray()
						fmt.Println("patch:%d", len(funcTokens))
						for _, funcToken := range funcTokens {
							fmt.Println("type:", funcToken.Type)
							if funcToken.Type == js.TokenJSBlock {

								compileContext.AddPatch(patchName, funcToken.Children)
								break
							}
						}
					}
				}

				jsScopeFile.Stream.AddToken(*token)
			}

			//jsScopeFile.Stream.Debug(0)
			jsScopeFile.IsLoaded = true
			fmt.Println("--------")
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
