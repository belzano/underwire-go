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
	"path/filepath"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

type TypeInfo struct {
	TypeName              string
	UnrealType            string
	IsEngineType          bool
	Layer                 string
	IsArray               bool
	IsTemplate            bool
	TemplateParamTypeName string
}

func Unique(typeInfos []TypeInfo) []TypeInfo {
	var result []TypeInfo
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

func RemoveEngineTypes(typeInfos []TypeInfo) []TypeInfo {
	var result []TypeInfo
	for _, typeInfo := range typeInfos {
		switch {
		case typeInfo.IsEngineType == true:
		default:
			result = append(result, typeInfo)
		}
	}
	return result
}

func Cleanup(typeInfos []TypeInfo) []TypeInfo {
	return slices.DeleteFunc(typeInfos, func(t TypeInfo) bool {
		return t.UnrealType == ""
	})
}

func Sort(typeInfos []TypeInfo) {
	slices.SortFunc(typeInfos, func(a, b TypeInfo) int {
		return cmp.Compare(a.UnrealType, b.UnrealType)
	})
}

func getUniqueExternalTypeInfos(typeInfos []TypeInfo) []TypeInfo {
	var result []TypeInfo

	for _, typeInfo := range typeInfos {
		switch {
		case typeInfo.IsEngineType == true:
		case slices.Contains(result, typeInfo):
		default:
			result = append(result, typeInfo)
		}
	}

	result = slices.DeleteFunc(result, func(t TypeInfo) bool {
		return t.UnrealType == ""
	})

	slices.SortFunc(result, func(a, b TypeInfo) int {
		return cmp.Compare(a.UnrealType, b.UnrealType)
	})

	return result
}

func getTypeInfo(ctx context.TemplateGenerationContext, prop *openapi3.SchemaRef) TypeInfo {
	if prop.Value.Type.Includes("array") {
		itemTypeInfo := getTypeInfo(ctx, prop.Value.Items)
		return TypeInfo{
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
		return TypeInfo{
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
			return TypeInfo{UnrealType: "FString", IsEngineType: true}
		case "integer":
			return TypeInfo{UnrealType: "int32", IsEngineType: true}
		case "number":
			return TypeInfo{UnrealType: "float", IsEngineType: true}
		case "boolean":
			return TypeInfo{UnrealType: "bool", IsEngineType: true}
		case "object":
			if prop.Value.AdditionalProperties.Schema != nil {
				itemTypeInfo := getTypeInfo(ctx, prop.Value.AdditionalProperties.Schema)
				return TypeInfo{
					UnrealType:   "TMap<FString, " + itemTypeInfo.UnrealType + ">",
					IsEngineType: itemTypeInfo.IsEngineType,
					Layer:        itemTypeInfo.Layer,
				}
			}

			return TypeInfo{UnrealType: "FJsonObjectWrapper", IsEngineType: true}
		}
	case prop.Value.Format == "date-time":
		return TypeInfo{UnrealType: "FDateTime", IsEngineType: true}
	}

	return TypeInfo{
		UnrealType:   "FString",
		IsEngineType: true,
	}
}
