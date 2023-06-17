package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xeipuuv/gojsonschema"
)

var (
	GoVersion  = runtime.Version()
	CommitHash = "n/a"
	BuildDate  = "n/a"
	OsArch     = runtime.GOOS + "/" + runtime.GOARCH
)

func main() {
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		showInfo()
		return
	}

	if flag.NArg() < 2 {
		showInfo()
		os.Exit(1)
	}

	schemaLoader, err := getLoader(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}

	schema, err := gojsonschema.NewSchema(schemaLoader)

	exitCode := 0
	for i := 1; i < flag.NArg(); i++ {
		documentLoader, err := getLoader(flag.Arg(i))
		if err != nil {
			log.Fatalln(err)
		}

		result, err := schema.Validate(documentLoader)
		if err != nil {
			log.Fatalln(err)
		}

		if result.Valid() {
			fmt.Printf("%s is valid\n", flag.Arg(i))
		} else {
			fmt.Printf("%s is not valid\n", flag.Arg(i))
			for _, err := range result.Errors() {
				fmt.Printf("- %s\n", err)
			}
			exitCode = 1
		}
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func showInfo() {
	fmt.Println("JSON Schema validator")
	fmt.Println("usage: validate-spec [path-to-schema] [path-to-specs]")
	fmt.Println("example: validate-spec spec.schema.json spec.json spec.min.json")
	fmt.Printf("- Go version: %s\n", GoVersion)
	fmt.Printf("- Git commit: %s\n", CommitHash)
	fmt.Printf("- Built:      %s\n", BuildDate)
	fmt.Printf("- OS/Arch:    %s\n", OsArch)
}

func getLoader(filename string) (gojsonschema.JSONLoader, error) {
	path, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return gojsonschema.NewBytesLoader(content), nil
}
