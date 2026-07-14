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

type Struct struct {
	Name          string
	Fields        []Field
	Dependencies  []TypeInfo
	IsAlias       bool
	AliasTypeInfo TypeInfo
}

func ExtractComponents(ctx context.TemplateGenerationContext, components *openapi3.Components) []Struct {
	var structs []Struct

	if components != nil && components.Schemas != nil {
		for name, schema := range components.Schemas {
			fields := extractFields(ctx, schema)

			if fields != nil {
				var depencies []TypeInfo
				for _, field := range fields {
					depencies = append(depencies, field.TypeInfo)
				}
				depencies = getUniqueExternalTypeInfos(depencies)

				structs = append(structs, Struct{
					Name:         name,
					Fields:       fields,
					Dependencies: depencies,
				})
				continue
			}

			if schema.Value.AdditionalProperties.Schema != nil {
				itemTypeInfo := getTypeInfo(ctx, schema.Value.AdditionalProperties.Schema)
				mapTypeInfo := TypeInfo{
					Kind:     KindMap,
					Layer:    itemTypeInfo.Layer,
					ItemType: &itemTypeInfo,
				}
				structs = append(structs, Struct{
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
