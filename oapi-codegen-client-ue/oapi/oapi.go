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
	"path/filepath"
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

type Field struct {
	Name         string
	TypeInfo     UnrealTypeInfo
	Dependencies []UnrealTypeInfo
}

type OapiStructTypeInfo struct {
	Name          string
	Fields        []Field
	Dependencies  []UnrealTypeInfo
	IsAlias       bool
	AliasTypeInfo UnrealTypeInfo
}

func ExtractComponents(ctx context.TemplateGenerationContext, components *openapi3.Components) []OapiStructTypeInfo {
	var structs []OapiStructTypeInfo

	if components != nil && components.Schemas != nil {
		for name, schema := range components.Schemas {
			fields := extractFields(ctx, schema)

			if fields != nil {
				var depencies []UnrealTypeInfo
				for _, field := range fields {
					depencies = append(depencies, field.TypeInfo)
				}
				depencies = getUniqueExternalTypeInfos(depencies)

				structs = append(structs, OapiStructTypeInfo{
					Name:         name,
					Fields:       fields,
					Dependencies: depencies,
				})
				continue
			}

			if schema.Value.AdditionalProperties.Schema != nil {
				itemTypeInfo := getUnrealTypeInfo(ctx, schema.Value.AdditionalProperties.Schema)
				mapTypeInfo := UnrealTypeInfo{
					UnrealType:   "TMap<FString, " + itemTypeInfo.UnrealType + ">",
					IsEngineType: itemTypeInfo.IsEngineType,
					Layer:        itemTypeInfo.Layer,
				}
				structs = append(structs, OapiStructTypeInfo{
					Name:          name,
					Fields:        nil,
					Dependencies:  nil,
					IsAlias:       true,
					AliasTypeInfo: mapTypeInfo,
				})
				continue
			}
		}
	}

	return structs
}

func extractFields(ctx context.TemplateGenerationContext, schema *openapi3.SchemaRef) []Field {
	var fields []Field

	for name, prop := range schema.Value.Properties {
		//log.Printf(" struct field '%s'", name)
		unrealTypeInfo := getUnrealTypeInfo(ctx, prop)
		fields = append(fields, Field{
			Name:     name,
			TypeInfo: unrealTypeInfo,
		})
	}

	return fields
}

type UnrealTypeInfo struct {
	TypeName              string
	UnrealType            string
	IsEngineType          bool
	Layer                 string
	IsArray               bool
	IsTemplate            bool
	TemplateParamTypeName string
}

func Unique(typeInfos []UnrealTypeInfo) []UnrealTypeInfo {
	var result []UnrealTypeInfo
	for _, typeInfo := range typeInfos {
		switch {
		case typeInfo.IsEngineType == true:
		case slices.Contains(result, typeInfo):
		default:
			result = append(result, typeInfo)
		}
	}
	return result
}

func RemoveEngineTypes(typeInfos []UnrealTypeInfo) []UnrealTypeInfo {
	var result []UnrealTypeInfo
	for _, typeInfo := range typeInfos {
		switch {
		case typeInfo.IsEngineType == true:
		default:
			result = append(result, typeInfo)
		}
	}
	return result
}

func Cleanup(typeInfos []UnrealTypeInfo) []UnrealTypeInfo {
	return slices.DeleteFunc(typeInfos, func(t UnrealTypeInfo) bool {
		return t.UnrealType == ""
	})
}

func Sort(typeInfos []UnrealTypeInfo) {
	slices.SortFunc(typeInfos, func(a, b UnrealTypeInfo) int {
		return cmp.Compare(a.UnrealType, b.UnrealType)
	})
}

func getUniqueExternalTypeInfos(typeInfos []UnrealTypeInfo) []UnrealTypeInfo {
	var result []UnrealTypeInfo

	for _, typeInfo := range typeInfos {
		switch {
		case typeInfo.IsEngineType == true:
		case slices.Contains(result, typeInfo):
		default:
			result = append(result, typeInfo)
		}
	}

	result = slices.DeleteFunc(result, func(t UnrealTypeInfo) bool {
		return t.UnrealType == ""
	})

	slices.SortFunc(result, func(a, b UnrealTypeInfo) int {
		return cmp.Compare(a.UnrealType, b.UnrealType)
	})

	return result
}

func getUnrealTypeInfo(ctx context.TemplateGenerationContext, prop *openapi3.SchemaRef) UnrealTypeInfo {
	if prop.Value.Type.Includes("array") {
		itemTypeInfo := getUnrealTypeInfo(ctx, prop.Value.Items)
		return UnrealTypeInfo{
			UnrealType:            "TArray<" + itemTypeInfo.UnrealType + ">",
			IsEngineType:          itemTypeInfo.IsEngineType,
			Layer:                 itemTypeInfo.Layer,
			IsArray:               true,
			IsTemplate:            true,
			TemplateParamTypeName: itemTypeInfo.TypeName,
		}
	}

	// $ref handling:
	// "#/components/schemas/User" -> "User"
	if prop.Ref != "" {
		schemaName := filepath.Base(prop.Ref)
		return UnrealTypeInfo{
			TypeName:     schemaName,
			UnrealType:   "F" + schemaName,
			IsEngineType: false,
			Layer:        ctx.Layer,
			IsArray:      false,
		}
	}

	switch {
	case prop.Value.Type.IsSingle():
		switch prop.Value.Type.Slice()[0] {
		case "string":
			return UnrealTypeInfo{UnrealType: "FString", IsEngineType: true}
		case "integer":
			return UnrealTypeInfo{UnrealType: "int32", IsEngineType: true}
		case "number":
			return UnrealTypeInfo{UnrealType: "float", IsEngineType: true}
		case "boolean":
			return UnrealTypeInfo{UnrealType: "bool", IsEngineType: true}
		case "object":
			if prop.Value.AdditionalProperties.Schema != nil {
				itemTypeInfo := getUnrealTypeInfo(ctx, prop.Value.AdditionalProperties.Schema)
				return UnrealTypeInfo{
					UnrealType:   "TMap<FString, " + itemTypeInfo.UnrealType + ">",
					IsEngineType: itemTypeInfo.IsEngineType,
					Layer:        itemTypeInfo.Layer,
				}
			}

			return UnrealTypeInfo{UnrealType: "FJsonObjectWrapper", IsEngineType: true}
		}
	case prop.Value.Format == "date-time":
		return UnrealTypeInfo{UnrealType: "FDateTime", IsEngineType: true}
	}

	return UnrealTypeInfo{
		UnrealType:   "FString",
		IsEngineType: true,
	}
}
