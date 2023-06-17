package export_to_openapi

import (
	"fmt"
	"strings"

	"github.com/alserom/tg-bot-api-spec/pkg/spec"
)

func info(as *spec.ApiSpec) map[string]interface{} {
	return map[string]interface{}{
		"title":   "Telegram Bot API",
		"version": as.GetVersion(),
		"description": strings.TrimSpace(fmt.Sprintf(`
This is a copy of the official [Telegram Bot API docs](https://core.telegram.org/bots/api) page converted to OpenAPI spec.

Note: Despite the fact that Telegram supports four ways of passing parameters in Bot API requests this spec describes only some of them per each method.
Methods without arguments describe only as 'GET' requests. Methods with 'InputFile' in arguments describe as 'POST' with content type 'multipart/form-data' & 'application/json'. Others - 'POST' with 'application/json' content type.

The Bot API is an HTTP-based interface created for developers keen on building bots for Telegram.
To learn how to create and set up a bot, please consult [Introduction to Bots](https://core.telegram.org/bots) and [Bot FAQ](https://core.telegram.org/bots/faq).

- Release date: %s
- Changelog: [%s](%s)
		`, as.GetReleaseDate(), as.GetLink(), as.GetLink())),
		"contact": map[string]interface{}{
			"name": "Generated with `tg-bot-api-spec` tool",
			"url":  "https://github.com/alserom/tg-bot-api-spec",
		},
		"license": map[string]interface{}{
			"name": "Licensed under the MIT License",
			"url":  "https://raw.githubusercontent.com/alserom/tg-bot-api-spec/main/LICENSE.md",
		},
	}
}

func externalDocs() map[string]interface{} {
	return map[string]interface{}{
		"description": "Telegram Bot API official page",
		"url":         "https://core.telegram.org/bots/api",
	}
}

func servers() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"url":         "https://api.telegram.org/bot{token}",
			"description": "Bot API Server",
			"variables": map[string]interface{}{
				"token": map[string]interface{}{
					"default": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				},
			},
		},
		{
			"url":         "{protocol}://{host}/bot{token}",
			"description": "Local Bot API Server",
			"variables": map[string]interface{}{
				"protocol": map[string]interface{}{
					"default": "http",
					"enum":    []string{"http", "https"},
				},
				"host": map[string]interface{}{
					"default": "localhost:8081",
				},
				"token": map[string]interface{}{
					"default": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				},
			},
		},
	}
}

func responses() map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"description": "Error",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
						"required": []string{
							"ok",
							"error_code",
						},
						"properties": map[string]interface{}{
							"ok": map[string]interface{}{
								"type":    "boolean",
								"default": false,
							},
							"error_code": map[string]interface{}{
								"type": "integer",
							},
							"description": map[string]interface{}{
								"type": "string",
							},
							"parameters": map[string]interface{}{
								"$ref": refToSchema("ResponseParameters"),
							},
						},
					},
				},
			},
		},
	}
}

func paths(as *spec.ApiSpec) (map[string]interface{}, error) {
	paths := make(map[string]interface{})

	for _, m := range as.GetMethods() {
		operation := map[string]interface{}{
			"tags":        [1]string{categoryToTag(m.GetCategory())},
			"description": m.GetDescription(),
			"summary":     "Describes `" + m.GetName() + "` method",
			"operationId": m.GetName(),
			"externalDocs": map[string]string{
				"description": "See official spec",
				"url":         m.GetLink(),
			},
		}

		var operationType string
		if len(m.GetArguments()) > 0 {
			operationType = "post"

			content, err := getRequestBodyContent(m)
			if err != nil {
				return nil, err
			}

			operation["requestBody"] = map[string]interface{}{"content": content}
		} else {
			operationType = "get"
		}

		resultSchema := map[string]interface{}{}
		if len(m.GetReturnTypes()) == 1 {
			setPropertyType(m.GetReturnTypes()[0], resultSchema)
		} else {
			oneOf := make([]map[string]interface{}, len(m.GetReturnTypes()))
			for i, dt := range m.GetReturnTypes() {
				oneOf[i] = make(map[string]interface{})
				setPropertyType(dt, oneOf[i])
			}
			resultSchema["oneOf"] = oneOf
		}

		operation["responses"] = map[string]interface{}{
			"200": map[string]interface{}{
				"description": "Success",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"required": []string{
								"ok",
								"result",
							},
							"properties": map[string]interface{}{
								"ok": map[string]interface{}{
									"type":    "boolean",
									"default": true,
								},
								"result": resultSchema,
								"description": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
			},
			"4XX":     map[string]interface{}{"$ref": "#/components/responses/error"},
			"5XX":     map[string]interface{}{"$ref": "#/components/responses/error"},
			"default": map[string]interface{}{"$ref": "#/components/responses/error"},
		}

		if m.GetName() == "setWebhook" {
			operation["callbacks"] = map[string]interface{}{
				"incomingUpdate": map[string]interface{}{
					"{$request.body#/url}": map[string]interface{}{
						"post": map[string]interface{}{
							"summary": "Webhook callback",
							"parameters": []map[string]interface{}{
								{
									"name":        "X-Telegram-Bot-Api-Secret-Token",
									"in":          "header",
									"description": "The header is useful to ensure that the request comes from a webhook set by you.",
									"schema":      map[string]interface{}{"type": "string"},
									"style":       "simple",
								},
							},
							"requestBody": map[string]interface{}{
								"required": true,
								"content": map[string]interface{}{
									"application/json": map[string]interface{}{
										"schema": map[string]interface{}{
											"$ref": refToSchema("Update"),
										},
									},
								},
							},
							"responses": map[string]interface{}{
								"200": map[string]interface{}{
									"description": "Your server returns this code if it accepts the callback. You can perform a request to the Bot API while sending an answer to the webhook. Use either application/json or application/x-www-form-urlencoded or multipart/form-data response content type for passing parameters. Specify the method to be invoked in the `method` parameter of the request. It's not possible to know that such a request was successful or get its result.",
								},
								"4XX": map[string]interface{}{
									"description": "Telegram will try to send a callback later. But it can give up after a reasonable amount of attempts.",
								},
								"5XX": map[string]interface{}{
									"description": "Telegram will try to send a callback later. But it can give up after a reasonable amount of attempts.",
								},
							},
						},
					},
				},
			}
		}

		paths["/"+m.GetName()] = map[string]interface{}{
			operationType: operation,
		}
	}

	return paths, nil
}

func getRequestBodyContent(m *spec.TgMethodSpec) (map[string]interface{}, error) {
	content := make(map[string]interface{})
	needMultipart, argsWithInputFile := isInputFileInDataTypes(m.GetArguments())
	var mediaType string
	if needMultipart {
		mediaType = "multipart/form-data"
	} else {
		mediaType = "application/json"
	}

	schema := map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
	}
	var required []string
	props := make(map[string]interface{})
	for _, a := range m.GetArguments() {
		if a.IsRequired() {
			required = append(required, a.GetName())
		}

		prop := map[string]interface{}{
			"description": a.GetDescription(),
		}

		dataTypes := a.GetDataTypes()
		if len(argsWithInputFile) > 0 {
			dataTypes = filterInputFileDataType(dataTypes)
		}

		if len(dataTypes) == 1 {
			setPropertyType(dataTypes[0], prop)
		} else {
			oneOf := make([]map[string]interface{}, len(dataTypes))
			for i, dt := range dataTypes {
				oneOf[i] = make(map[string]interface{})
				setPropertyType(dt, oneOf[i])
			}
			prop["oneOf"] = oneOf
		}

		if len(dataTypes) > 0 {
			props[a.GetName()] = prop
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	schema["properties"] = props

	if len(argsWithInputFile) > 0 {
		content["application/json"] = map[string]interface{}{"schema": schema}
		schema2 := make(map[string]interface{})
		for k, v := range schema {
			schema2[k] = v
		}
		props2 := make(map[string]interface{})
		for k, v := range props {
			props2[k] = v
		}

		for argName, propData := range argsWithInputFile {
			props2[argName] = propData
		}
		schema2["properties"] = props2
		content["multipart/form-data"] = map[string]interface{}{"schema": schema2}
	} else {
		content[mediaType] = map[string]interface{}{"schema": schema}
	}

	return content, nil
}

func isInputFileInDataTypes(args []*spec.TgMethodSpecArgument) (bool, map[string]interface{}) {
	argNames := make(map[string]interface{})
	for _, a := range args {
		if len(a.GetDataTypes()) == 1 && a.GetDataTypes()[0].GetDefinition() == "InputFile" && a.IsRequired() {
			return true, make(map[string]interface{})
		}

		for _, dt := range a.GetDataTypes() {
			if dt.GetDefinition() == "InputFile" {
				argNames[a.GetName()] = map[string]interface{}{
					"description": a.GetDescription(),
					"$ref":        refToSchema("InputFile"),
				}
			}
		}
	}

	return len(argNames) > 0, argNames
}

func filterInputFileDataType(dataTypes []spec.DataTypeDefinition) []spec.DataTypeDefinition {
	var filtered []spec.DataTypeDefinition
	for _, dt := range dataTypes {
		if dt.GetDefinition() != "InputFile" {
			filtered = append(filtered, dt)
		}
	}

	return filtered
}

func schemas(as *spec.ApiSpec) (map[string]interface{}, error) {
	schemas := make(map[string]interface{})

	for _, t := range as.GetTypes() {
		obj := map[string]interface{}{
			"type":                 "object",
			"additionalProperties": false,
			"description":          t.GetDescription(),
			"externalDocs": map[string]string{
				"description": "See official spec",
				"url":         t.GetLink(),
			},
		}

		if t.GetName() == "InputFile" {
			obj["type"] = "string"
			obj["format"] = "binary"
			delete(obj, "additionalProperties")
			schemas[t.GetName()] = obj
			continue
		}

		if len(t.GetChildren()) > 0 {
			oneOf := make([]map[string]string, len(t.GetChildren()))
			discriminatorMapping := make(map[string]string)
			discriminatorPropName := ""
			dpnCheck := make(map[string]bool)
			for i, child := range t.GetChildren() {
				oneOf[i] = map[string]string{"$ref": refToSchema(child.GetName())}
				for _, p := range child.GetProperties() {
					if p.GetPredefinedValue() != nil {
						discriminatorMapping[string(*p.GetPredefinedValue())] = refToSchema(child.GetName())
						discriminatorPropName = p.GetName()
						dpnCheck[p.GetName()] = true
						break
					}
				}
			}

			obj["oneOf"] = oneOf

			if len(oneOf) == len(discriminatorMapping) && len(dpnCheck) == 1 {
				obj["discriminator"] = map[string]interface{}{
					"propertyName": discriminatorPropName,
					"mapping":      discriminatorMapping,
				}
			}
		}

		if len(t.GetProperties()) > 0 {
			var required []string
			props := make(map[string]interface{})
			for _, p := range t.GetProperties() {
				prop := map[string]interface{}{
					"description": p.GetDescription(),
				}

				if p.GetPredefinedValue() != nil {
					prop["default"] = string(*p.GetPredefinedValue())
				}

				if !p.IsOptional() {
					required = append(required, p.GetName())
				}

				if len(p.GetDataTypes()) == 1 {
					setPropertyType(p.GetDataTypes()[0], prop)
				} else {
					oneOf := make([]map[string]interface{}, len(p.GetDataTypes()))
					for i, dt := range p.GetDataTypes() {
						oneOf[i] = make(map[string]interface{})
						setPropertyType(dt, oneOf[i])
					}
					prop["oneOf"] = oneOf
				}

				props[p.GetName()] = prop
			}

			if len(required) > 0 {
				obj["required"] = required
			}

			obj["properties"] = props
		}

		schemas[t.GetName()] = obj
	}

	return schemas, nil
}

func refToSchema(name string) string {
	return "#/components/schemas/" + name
}

func setPropertyType(dtDef spec.DataTypeDefinition, prop map[string]interface{}) {
	switch dt := dtDef.(type) {
	case *spec.ObjectDataType:
		prop["$ref"] = refToSchema(dt.GetRef().GetName())
	case *spec.ArrayDataType:
		prop["type"] = "array"
		items := make(map[string]interface{})
		if len(dt.GetElementDataTypes()) == 1 {
			setPropertyType(dt.GetElementDataTypes()[0], items)
		} else {
			oneOf := make([]map[string]interface{}, len(dt.GetElementDataTypes()))
			for i, ai := range dt.GetElementDataTypes() {
				oneOf[i] = make(map[string]interface{})
				setPropertyType(ai, oneOf[i])
			}
			items["oneOf"] = oneOf
		}

		prop["items"] = items
	case *spec.ScalarDataType:
		dtTitle, dtFormat := scalarTypeFromDefinition(*dt)
		prop["type"] = dtTitle
		if dtFormat != "" {
			prop["format"] = dtFormat
		}
	}
}

func scalarTypeFromDefinition(dt spec.ScalarDataType) (string, string) {
	switch dt.GetDefinition() {
	case "float":
		return "number", "float"
	case "int32":
		return "integer", "int32"
	case "int64":
		return "integer", "int64"
	}

	return strings.ToLower(dt.GetDefinition()), ""
}

func categoryToTag(category string) string {
	return strings.ReplaceAll(category, "-", " ")
}
