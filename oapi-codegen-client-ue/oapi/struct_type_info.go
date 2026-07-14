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
	"oapi-codegen-client-ue/context"

	"github.com/getkin/kin-openapi/openapi3"
)

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
