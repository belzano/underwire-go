// Copyright (c) 2026 Guillaume Evrard
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

func main() {

	spec := *flag.String("spec", "../jij-service/api/openapi.yaml", "source openapi contract")
	outputDir := *flag.String("outdir", "./output", "output directory")

	swagger, err := openapi3.NewLoader().LoadFromFile(spec)
	if err != nil {
		log.Fatalf("Failed to load OpenAPI definition: %v", err)
	}

	layer := "serviceclient"
	for _, dir := range [...]string{outputDir, outputDir + "/" + layer} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("failed to create output dir: %v", err)
		}
	}

	if swagger.Components != nil && swagger.Components.Schemas != nil {
		for name, schema := range swagger.Components.Schemas {
			generateStruct(outputDir, layer, name, schema, swagger)
		}
	} else {
		log.Fatal("No schema found")
	}

	if swagger.Paths != nil && swagger.Paths.Len() > 0 {
		generateServiceClient(outputDir, layer, swagger.Paths, swagger)
	} else {
		log.Print("No path found")
	}

	generateServiceClientHelper(outputDir, layer)

	log.Println("generation complete")
}

func generateServiceClientHelper(outputDir string, layer string) {
	data := struct {
		Layer string
	}{
		Layer: layer,
	}
	gen := map[string]string{
		"templates/serviceclienthelper.h.tmpl":   "ServiceClientHelper.h",
		"templates/serviceclienthelper.cpp.tmpl": "ServiceClientHelper.cpp",
	}
	for tmplFilePath, outFilePath := range gen {
		tmpl, err := template.ParseFiles(tmplFilePath)
		if err != nil {
			log.Fatalf("failed to load template: %v", err)
		}

		outputFile := filepath.Join(outputDir+"/"+layer, outFilePath)
		file, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("failed to create file %s: %v", outputFile, err)
		}
		defer file.Close()

		if err := tmpl.Execute(file, data); err != nil {
			log.Fatalf("failure during template execution: %v", err)
		}
	}
}

func generateStruct(outputDir string, layer string, name string, schema *openapi3.SchemaRef, swagger *openapi3.T) {
	data := struct {
		Layer  string
		Name   string
		Fields []Field
	}{
		Layer:  layer,
		Name:   name,
		Fields: getFields(schema, swagger),
	}

	tmpl, err := template.ParseFiles("templates/struct.tmpl")
	if err != nil {
		log.Fatalf("failed to load template: %v", err)
	}

	outputFile := filepath.Join(outputDir+"/"+layer, name+".h")
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("failed to create file %s: %v", outputFile, err)
	}
	defer file.Close()

	// 4. Exécuter le template
	if err := tmpl.Execute(file, data); err != nil {
		log.Fatalf("failure during template execution: %v", err)
	}
}

type Field struct {
	Layer   string
	Type    string
	Name    string
	IsArray bool
}

func getFields(schema *openapi3.SchemaRef, swagger *openapi3.T) []Field {
	var fields []Field

	for name, prop := range schema.Value.Properties {
		fieldType, isArray := getUnrealType(prop, swagger)
		fields = append(fields, Field{
			Type:    fieldType,
			Name:    name,
			IsArray: isArray,
		})
	}

	return fields
}

func getUnrealType(prop *openapi3.SchemaRef, swagger *openapi3.T) (string, bool) {
	if prop.Value.Type.Includes("array") {
		itemType, _ := getUnrealType(prop.Value.Items, swagger)
		return "TArray<" + itemType + ">", true
	}

	// $ref handling:
	// "#/components/schemas/User" -> "User"
	if prop.Ref != "" {
		schemaName := filepath.Base(prop.Ref)
		return "F" + schemaName, false
	}

	switch {
	case prop.Value.Type.IsSingle():
		switch prop.Value.Type.Slice()[0] {
		case "string":
			return "FString", false
		case "integer":
			return "int32", false
		case "number":
			return "float", false
		case "boolean":
			return "bool", false
		}
	case prop.Value.Format == "date-time":
		return "FDateTime", false
	}

	return "FString", false // Fallback
}

func generateServiceClient(outputDir string, layer string, paths *openapi3.Paths, swagger *openapi3.T) {
	data := struct {
		Layer             string
		ServiceClientName string
		Endpoints         []ServiceEndpoint
	}{
		Layer:             layer,
		ServiceClientName: "ServiceClient",
		Endpoints:         getServiceEndpoints(paths, swagger),
	}

	gen := map[string]string{
		"templates/serviceclient.h.tmpl":   "ServiceClient.h",
		"templates/serviceclient.cpp.tmpl": "ServiceClient.cpp",
	}
	for tmplFilePath, outFilePath := range gen {
		tmpl, err := template.ParseFiles(tmplFilePath)
		if err != nil {
			log.Fatalf("failed to load template: %v", err)
		}

		outputFile := filepath.Join(outputDir+"/"+layer, outFilePath)
		file, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("failed to create file %s: %v", outputFile, err)
		}
		defer file.Close()

		if err := tmpl.Execute(file, data); err != nil {
			log.Fatalf("failure during template execution: %v", err)
		}
	}
}

type ServiceEndpoint struct {
	Name             string
	Verb             string
	PathPrintfStyle  string
	QueryParameters  []Field
	QueryBodyType    string
	ResponseBodyType string
}

func getServiceEndpoints(paths *openapi3.Paths, swagger *openapi3.T) []ServiceEndpoint {
	var endpoints []ServiceEndpoint

	for path, pathItem := range paths.Map() {
		for method, operation := range pathItem.Operations() {
			endpoint := getServiceEndpoint(method, path, operation, swagger)
			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints
}

func getServiceEndpoint(verb string, path string, operation *openapi3.Operation, swagger *openapi3.T) ServiceEndpoint {

	var queryParameters []Field

	var parameters []*openapi3.ParameterRef = operation.Parameters
	for _, parameter := range parameters {
		switch parameter.Value.In {
		case "path":
			unrealType, _ := getUnrealType(parameter.Value.Schema, swagger)
			field := Field{
				Name: parameter.Value.Name,
				Type: unrealType,
			}
			queryParameters = append(queryParameters, field)
		}
	}

	var queryBodyType string
	if operation.RequestBody != nil {
		mediaType := operation.RequestBody.Value.Content.Get("application/json")
		schemaRef := mediaType.Schema
		queryBodyType, _ = getUnrealType(schemaRef, swagger)
	}

	var responseBodyType string
	for _, responseRef := range operation.Responses.Map() {

		mediaType := responseRef.Value.Content.Get("application/json")
		if mediaType == nil {
			continue
		}
		schemaRef := mediaType.Schema
		responseBodyType, _ = getUnrealType(schemaRef, swagger)
	}

	re := regexp.MustCompile(`\{[^{}]*\}`)
	pathNoParams := re.ReplaceAllString(path, "")
	name := toPascalCase(verb + "/" + pathNoParams)

	pathPrintfStyle := re.ReplaceAllString(path, "%s")

	endpoint := ServiceEndpoint{
		Name:             name,
		Verb:             verb,
		PathPrintfStyle:  pathPrintfStyle,
		QueryParameters:  queryParameters,
		QueryBodyType:    queryBodyType,
		ResponseBodyType: responseBodyType,
	}

	return endpoint
}

func toPascalCase(input string) string {
	// Match any non-alphanumeric character as a delimiter
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	words := reg.Split(input, -1)

	for i, word := range words {
		if len(word) > 0 {
			// Title case each word individually
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, "")
}
