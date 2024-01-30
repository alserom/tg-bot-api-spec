package scrape

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"

	"github.com/alserom/tg-bot-api-spec/pkg/spec"
)

const url string = "https://core.telegram.org"
const url_api_doc string = url + "/bots/api"

type Scraper struct {
	doc *goquery.Document
}

type tgVersionSpec struct {
	version     string
	releaseDate string
	link        string
}

type deferredTgTypeSpecChild struct {
	childName string
	parent    *spec.TgTypeSpec
}

type helper struct {
	declareDataType func(definition string) spec.DataTypeDefinition
}

func createHelper(as *spec.ApiSpec) helper {
	return helper{
		declareDataType: func(definition string) spec.DataTypeDefinition {
			return as.DeclareDataType(definition)
		},
	}
}

func createScraper(r io.Reader) (*Scraper, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	return &Scraper{doc: doc}, nil
}

func NewScraper() (*Scraper, error) {
	res, err := http.Get(url_api_doc)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Get \"%s\": %d %s", url_api_doc, res.StatusCode, res.Status))
	}

	return createScraper(res.Body)
}

func NewFileScraper(path string) (*Scraper, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return createScraper(f)
}

func (s *Scraper) FillApiSpec(as *spec.ApiSpec) error {
	if s.doc == nil {
		return errors.New("document missed, nothing to scrape")
	}

	childToParent := make(map[string]*spec.TgTypeSpec)

	for scrapeItem := range scrape(s.doc, createHelper(as)) {
		switch item := scrapeItem.(type) {
		case *tgVersionSpec:
			if as.GetVersion() == "" {
				as.SetVersion(item.version)
				as.SetReleaseDate(item.releaseDate)
				as.SetLink(item.link)
			}
		case *spec.TgTypeSpec:
			key := item.GetName()
			if parent, exists := childToParent[key]; exists {
				item.SetParent(parent)
				parent.AddChild(item)
				delete(childToParent, key)
			}

			as.AddType(item)
		case *spec.TgMethodSpec:
			as.AddMethod(item)
		case *deferredTgTypeSpecChild:
			childToParent[item.childName] = item.parent
		case error:
			return item
		}
	}

	if len(childToParent) > 0 {
		for childName, parent := range childToParent {
			if child, ok := as.GetType(childName); ok {
				child.SetParent(parent)
				parent.AddChild(child)
				delete(childToParent, childName)
			}
		}
	}

	if len(childToParent) > 0 {
		msg := "some types were not added to the spec:"
		for childName := range childToParent {
			msg += "\n- " + childName
		}
		return errors.New(msg)
	}

	return nil
}
