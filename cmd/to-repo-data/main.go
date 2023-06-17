package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alserom/tg-bot-api-spec/internal/datasource/scrape"
	export_to_openapi "github.com/alserom/tg-bot-api-spec/internal/export/openapi"
	export_to_json "github.com/alserom/tg-bot-api-spec/pkg/export/json"
	"github.com/alserom/tg-bot-api-spec/pkg/spec"
)

var (
	GoVersion  = runtime.Version()
	CommitHash = "n/a"
	BuildDate  = "n/a"
	OsArch     = runtime.GOOS + "/" + runtime.GOARCH
)

type Exporter interface {
	Export(filename string) error
}

func main() {
	source := flag.String(
		"source",
		"",
		"Path to '*.html' file which can be a data source for scraping. If empty - scraping https://core.telegram.org/bots/api.",
	)
	dir := flag.String("dir", "", "Path to the output directory")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help || flag.Arg(0) == "help" {
		showInfo()
		return
	}

	err := execute(*source, *dir)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func execute(source, dir string) error {
	fail := true
	out, isCreated, err := prepareDir(dir)
	if err != nil {
		return err
	}
	if isCreated {
		defer removeCreatedDirOnFail(out, &fail)
	}

	fmt.Println("initializing data source...")
	datasource, err := getDatasource(source)
	if err != nil {
		return err
	}

	fmt.Println("creating specification...")
	spec, err := spec.NewApiSpec(datasource)
	if err != nil {
		return err
	}

	fmt.Printf("Bot API v%s created\n", spec.GetVersion())

	fmt.Println("creating exporters...")

	jsonExporter, err := export_to_json.NewApiSpecExporter(*spec)
	if err != nil {
		return err
	}

	openapiExporter, err := export_to_openapi.NewOpenapiExporter(*spec)
	if err != nil {
		return err
	}

	fmt.Println("exporting...")

	err = jsonExporter.Export(out)
	if err != nil {
		return err
	}

	err = openapiExporter.Export(out)
	if err != nil {
		return err
	}

	outVersion := out + "/version.json"
	err = ioutil.WriteFile(outVersion, []byte(fmt.Sprintf(`{"version":"%s"}`, spec.GetVersion())), 0644)
	fmt.Println("saving: " + outVersion)
	if err != nil {
		return err
	}

	fail = false

	return nil
}

func getDatasource(source string) (spec.DataSource, error) {
	if source == "" {
		return scrape.NewScraper()
	}

	path, err := filepath.Abs(source)
	if err != nil {
		return nil, err
	}

	return scrape.NewFileScraper(path)
}

func prepareDir(dir string) (string, bool, error) {
	outPath, err := filepath.Abs(dir)
	if err != nil {
		return "", false, err
	}

	fileInfo, err := os.Stat(outPath)
	if err == nil && fileInfo.IsDir() {
		return outPath, false, nil
	}

	err = os.Mkdir(outPath, os.ModePerm)
	if err != nil {
		return "", false, err
	}

	return outPath, true, nil
}

func showInfo() {
	fmt.Println("Telegram Bot API spec exporter")
	fmt.Printf("- Go version: %s\n", GoVersion)
	fmt.Printf("- Git commit: %s\n", CommitHash)
	fmt.Printf("- Built:      %s\n", BuildDate)
	fmt.Printf("- OS/Arch:    %s\n", OsArch)
	fmt.Println("usage: to-repo-data [flags]")
	flag.PrintDefaults()
}

func removeCreatedDirOnFail(path string, fail *bool) {
	if !*fail {
		return
	}

	os.RemoveAll(path)
}
