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

package oapi

import (
	"cmp"
	"oapi-codegen-client-ue/context"
	"oapi-codegen-client-ue/utils"
	"regexp"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

type ServiceEndpoint struct {
	Name             string
	Verb             string
	PathPrintfStyle  string
	QueryParameters  []Field
	QueryBodyType    UnrealTypeInfo
	ResponseBodyType UnrealTypeInfo
}

func (e *ServiceEndpoint) GetDependantTypes() []UnrealTypeInfo {
	var typeInfos []UnrealTypeInfo

	for _, queryParam := range e.QueryParameters {
		typeInfos = append(typeInfos, queryParam.TypeInfo)
	}
	typeInfos = append(typeInfos, e.QueryBodyType)
	typeInfos = append(typeInfos, e.ResponseBodyType)

	return typeInfos
}

func ExtractServiceEndpoints(ctx context.TemplateGenerationContext, paths *openapi3.Paths) []ServiceEndpoint {
	var endpoints []ServiceEndpoint

	if paths == nil {
		return endpoints
	}

	for path, pathItem := range paths.Map() {
		for method, operation := range pathItem.Operations() {
			endpoint := getServiceEndpoint(ctx, method, path, operation)
			endpoints = append(endpoints, endpoint)
		}
	}

	slices.SortFunc(endpoints, func(a, b ServiceEndpoint) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return endpoints
}

func getServiceEndpoint(ctx context.TemplateGenerationContext, verb string, path string, operation *openapi3.Operation) ServiceEndpoint {

	var queryParameters []Field

	var parameters []*openapi3.ParameterRef = operation.Parameters
	for _, parameter := range parameters {
		switch parameter.Value.In {
		case "path":
			typeInfo := getUnrealTypeInfo(ctx, parameter.Value.Schema)
			field := Field{
				Name:     parameter.Value.Name,
				TypeInfo: typeInfo,
			}
			queryParameters = append(queryParameters, field)
		}
	}

	var queryBodyType UnrealTypeInfo
	if operation.RequestBody != nil {
		mediaType := operation.RequestBody.Value.Content.Get("application/json")
		schemaRef := mediaType.Schema
		queryBodyType = getUnrealTypeInfo(ctx, schemaRef)
	}

	var responseBodyType UnrealTypeInfo
	for _, responseRef := range operation.Responses.Map() {

		mediaType := responseRef.Value.Content.Get("application/json")
		if mediaType == nil {
			continue
		}
		schemaRef := mediaType.Schema
		responseBodyType = getUnrealTypeInfo(ctx, schemaRef)
	}

	re := regexp.MustCompile(`\{[^{}]*\}`)
	pathNoParams := re.ReplaceAllString(path, "")
	name := utils.ToPascalCase(pathNoParams + "/" + verb)

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
