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
