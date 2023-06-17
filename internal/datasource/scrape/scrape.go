package scrape

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/alserom/tg-bot-api-spec/pkg/spec"
)

const url_changelog string = url + "/bots/api-changelog"

func scrape(doc *goquery.Document, h helper) chan interface{} {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		var category string
		var item interface{}

		doc.Find("#dev_page_content").Children().EachWithBreak(func(i int, s *goquery.Selection) bool {
			nodeName := goquery.NodeName(s)
			switch nodeName {
			case "h3", "h4":
				anchor := s.Find("a.anchor")
				anchorName, exists := anchor.Attr("name")
				if !exists {
					ch <- errors.New(fmt.Sprintf("scraping error: detected node %s without anchor", nodeName))
					return false
				}

				if item != nil {
					ch <- item
				}

				if nodeName == "h3" {
					category = anchorName
					item = nil
				} else {
					var err error
					item, err = newSpecItem(category, anchorName, s.Text(), anchor.AttrOr("href", ""))
					if err != nil {
						ch <- errors.New("scraping error: can't create new item. error: " + err.Error())
						return false
					}
				}
			default:
				switch specItem := item.(type) {
				case *tgVersionSpec:
					fillTgVersionSpec(specItem, nodeName, s)
				case *spec.TgTypeSpec:
					err := fillTgTypeSpec(specItem, nodeName, s, ch, h)
					if err != nil {
						ch <- err
						return false
					}
				case *spec.TgMethodSpec:
					err := fillTgMethodSpec(specItem, nodeName, s, h)
					if err != nil {
						ch <- err
						return false
					}
				}
			}

			return true
		})

		if item != nil {
			ch <- item
		}
	}()

	return ch
}

func newSpecItem(category, anchorName, name, link string) (interface{}, error) {
	if category == "recent-changes" {
		link = hrefToLink(link, url_changelog)

		return &tgVersionSpec{
			releaseDate: name,
			link:        link,
		}, nil
	} else if name != "" && !strings.Contains(anchorName, "-") {

		link = hrefToLink(link, url_api_doc)

		if name[0] == strings.ToUpper(name)[0] {
			return spec.NewTgTypeSpec(category, name, link)
		} else {
			return spec.NewTgMethodSpec(category, name, link)
		}
	}

	return nil, nil
}

func hrefToLink(href, hashPrefix string) string {
	switch {
	case strings.HasPrefix(href, "#"):
		return hashPrefix + href
	case strings.HasPrefix(href, "/"):
		return url + href
	}

	return href
}

func fillTgVersionSpec(item *tgVersionSpec, nodeName string, s *goquery.Selection) {
	if nodeName == "p" && item.version == "" {
		item.version = s.Text()
		matches := regexp.MustCompile(`(?:Bot API )([\d.]*)?`).FindStringSubmatch(item.version)
		if len(matches) == 2 {
			item.version = matches[1]
		}
	}
}

func fillTgTypeSpec(item *spec.TgTypeSpec, nodeName string, s *goquery.Selection, ch chan interface{}, h helper) error {
	var err error
	switch nodeName {
	case "p":
		item.SetDescription(concatDescription(item.GetDescription(), prepareDescription(s)))
	case "ul":
		subtypes := strings.Split(strings.TrimSpace(s.Text()), "\n")
		for _, childName := range subtypes {
			ch <- &deferredTgTypeSpecChild{
				childName: childName,
				parent:    item,
			}
		}

		item.SetDescription(concatDescription(item.GetDescription(), prepareDescription(s)))
	case "table":
		s.Find("tbody > tr").EachWithBreak(func(row int, tr *goquery.Selection) bool {
			var property *spec.TgTypeSpecProperty
			tr.Find("td").EachWithBreak(func(column int, td *goquery.Selection) bool {
				switch column {
				case 0:
					property, err = spec.NewTgTypeSpecProperty(td.Text())
					if err != nil {
						err = errors.New("can't create new TgTypeSpecProperty, error: " + err.Error())
						return false
					}
				case 1:
					text := td.Text()
					if strings.Contains(text, "Integer") && strings.Contains(td.Next().Text(), "64-bit integer") {
						text = strings.ReplaceAll(text, "Integer", "Integer64")
					}

					types := extractTypes(text, h)
					if len(types) == 0 {
						err = errors.New(fmt.Sprintf("scraping error: can't parse data type for property '%s' of object '%s'", property.GetName(), item.GetName()))

						return false
					}
					for _, t := range types {
						property.AddDataType(t)
					}
				case 2:
					property.SetOptional(strings.HasPrefix(td.Text(), "Optional."))

					if !property.IsOptional() {
						html, _ := td.Html()
						property.SetPredefinedValue(extractPredefinedValue(html))
					}

					property.SetDescription(concatDescription(property.GetDescription(), prepareDescription(td)))
				default:
					err = errors.New(fmt.Sprintf("scraping error: can't parse properties of object '%s', too many columns", item.GetName()))

					return false
				}

				return true
			})

			if err != nil {
				return false
			} else if property == nil {
				err = errors.New(fmt.Sprintf("scraping error: expecting property for object %s, but it's missed", item.GetName()))

				return false
			}

			item.AddProperty(property)

			return true
		})
	}

	return err
}

func fillTgMethodSpec(item *spec.TgMethodSpec, nodeName string, s *goquery.Selection, h helper) error {
	var err error
	switch nodeName {
	case "p":
		if item.GetDescription() == "" {
			for _, returnType := range extractReturnTypes(s.Text(), h) {
				item.AddReturnType(returnType)
			}
		}

		item.SetDescription(concatDescription(item.GetDescription(), prepareDescription(s)))
	case "table":
		s.Find("tbody > tr").EachWithBreak(func(row int, tr *goquery.Selection) bool {
			var argument *spec.TgMethodSpecArgument
			tr.Find("td").EachWithBreak(func(column int, td *goquery.Selection) bool {
				switch column {
				case 0:
					argument, err = spec.NewTgMethodSpecArgument(td.Text())
					if err != nil {
						err = errors.New("can't create new TgMethodSpecArgument, error: " + err.Error())
						return false
					}
				case 1:
					text := td.Text()
					if strings.Contains(text, "Integer") && strings.Contains(td.Next().Next().Text(), "64-bit integer") {
						text = strings.ReplaceAll(text, "Integer", "Integer64")
					}

					types := extractTypes(text, h)
					if len(types) == 0 {
						err = errors.New(fmt.Sprintf("scraping error: can't parse data type for argument '%s' of method '%s'", argument.GetName(), item.GetName()))

						return false
					}
					for _, t := range types {
						argument.AddDataType(t)
					}
				case 2:
					argument.SetRequired(td.Text() == "Yes")
				case 3:
					argument.SetDescription(concatDescription(argument.GetDescription(), prepareDescription(td)))
				default:
					err = errors.New(fmt.Sprintf("scraping error: can't parse arguments of method '%s', too many columns", item.GetName()))

					return false
				}

				return true
			})

			if err != nil {
				return false
			} else if argument == nil {
				err = errors.New(fmt.Sprintf("scraping error: expecting argument for method %s, but it's missed", item.GetName()))

				return false
			}

			item.AddArgument(argument)

			return true
		})
	}

	return err
}

func extractPredefinedValue(html string) *spec.TgTypeSpecPropertyValue {
	matches := regexp.MustCompile(`(?i)(?:always “|must be <em>)(\b[A-Z].*?\b)?`).FindStringSubmatch(html)
	if len(matches) == 2 {
		var value spec.TgTypeSpecPropertyValue = spec.TgTypeSpecPropertyValue(matches[1])

		return &value
	}

	return nil
}

func extractReturnTypes(text string, h helper) []spec.DataTypeDefinition {
	var types []spec.DataTypeDefinition

	matches := regexp.MustCompile(`(?i)(?:on success,|returns)([^.]*)(?:on success)?`).FindStringSubmatch(text)
	if len(matches) < 2 {
		matches = regexp.MustCompile(`(?i)(?:An)([^.]*)(?:is returned)`).FindStringSubmatch(text)
	}

	if len(matches) < 2 {
		return types
	}

	matches[1] = strings.ReplaceAll(matches[1], "Array", "array")
	typesList := regexp.MustCompile(`\b[A-Z].*?\b`).FindAllString(matches[1], -1)
	if len(typesList) == 0 {
		return types
	}

	join, prefix := " or ", ""
	if regexp.MustCompile(`(?i)(?:array of )`).MatchString(matches[1]) {
		join, prefix = ", ", "Array of "
	}

	return extractTypes(prefix+strings.Join(typesList, join), h)
}

func extractTypes(text string, h helper) []spec.DataTypeDefinition {
	oneOfTypes := strings.Split(text, " or ")

	var types []spec.DataTypeDefinition

	for _, unparsedType := range oneOfTypes {
		typeDef := extractTypeDef(unparsedType)
		if typeDef != "" {
			types = append(types, h.declareDataType(typeDef))
		}
	}

	return types
}

func extractTypeDef(text string) string {
	var def string
	text = strings.ReplaceAll(text, " and ", ", ")

	if strings.HasPrefix(text, "Array of ") {
		def = "array"
		text = strings.TrimPrefix(text, "Array of ")
		itemsDef := extractTypeDef(text)
		if strings.HasPrefix(itemsDef, "_anyOf") {
			itemsDef = strings.ReplaceAll(itemsDef, "_anyOf", "")
		}
		if itemsDef != "" {
			def += "<" + itemsDef + ">"
		}
	} else if strings.Contains(text, ", ") {
		def = "_anyOf"
		anyOfTypes := strings.Split(text, ",")
		var extractedAnyOfTypes []string
		for _, unparsedType := range anyOfTypes {
			typeDef := extractTypeDef(unparsedType)
			if typeDef != "" {
				extractedAnyOfTypes = append(extractedAnyOfTypes, typeDef)
			}
		}
		def += strings.Join(extractedAnyOfTypes, "|")

	} else {
		def = fixTypeDef(strings.TrimSpace(text))
	}

	return def
}

func fixTypeDef(typeDef string) string {
	switch typeDef {
	case "True", "False", "Bool", "Boolean":
		return "boolean"
	case "Float number", "Float":
		return "float"
	case "Int", "Integer":
		return "int32"
	case "Integer64":
		return "int64"
	case "String":
		return "string"
	case "Messages":
		return "Message"
	}

	return typeDef
}

func concatDescription(current, new string) string {
	nl := "\n"

	lines := strings.Split(current, "\n")
	var lastLine string
	if len(lines) > 1 {
		lastLine = lines[len(lines)-1]
	} else {
		lastLine = current
	}

	if strings.HasPrefix(new, "- ") || strings.HasPrefix(lastLine, "- ") {
		nl = "\n\n"
	}

	return strings.TrimSpace(current + nl + new)
}

func prepareDescription(s *goquery.Selection) string {
	pref := ""
	if goquery.NodeName(s) == "ul" {
		pref = "- "
	}

	s.Find("a[href]").Each(func(i int, a *goquery.Selection) {
		title := a.Text()
		link := hrefToLink(a.AttrOr("href", ""), url_api_doc)

		if title != link && link != "" {
			a.SetText(fmt.Sprintf("%s[%s](%s)", pref, title, link))
		}
	})

	return strings.TrimSpace(s.Text())
}
