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

// Kind is a framework-agnostic classification of an OpenAPI schema, meant to
// let callers (like the tmpl package) derive their own target-language type
// spelling without needing to know anything about Unreal/C++.
type Kind string

const (
	KindString   Kind = "string"
	KindInteger  Kind = "integer"
	KindNumber   Kind = "number"
	KindBoolean  Kind = "boolean"
	KindDateTime Kind = "datetime"
	KindObject   Kind = "object"
	KindArray    Kind = "array"
	KindMap      Kind = "map"
	KindRef      Kind = "ref"
)

func (k Kind) IsArray() bool {
	return k == KindArray
}

// IsPrimitive reports whether Kind alone is always a built-in Unreal type
// needing no #include. It only holds for the non-container kinds: an array
// or map's own primitiveness actually depends on what it contains, so this
// intentionally simplifies TArray</TMap<-of-primitives to "not primitive" —
// a corner case this generator's current schemas never hit.
func (k Kind) IsPrimitive() bool {
	switch k {
	case KindArray, KindMap, KindRef:
		return false
	default:
		return true
	}
}

// TypeInfo is a framework-agnostic description of a schema's shape. It
// carries no target-language spelling (see tmpl.TypeInfo for that) — just
// enough structure (Kind, the referenced schema's TypeName, and, for
// containers, the element ItemType) for a renderer to derive one.
type TypeInfo struct {
	Kind     Kind
	TypeName string
	Layer    string
	ItemType *TypeInfo
}

func (t TypeInfo) equal(other TypeInfo) bool {
	if t.Kind != other.Kind || t.TypeName != other.TypeName || t.Layer != other.Layer {
		return false
	}
	if (t.ItemType == nil) != (other.ItemType == nil) {
		return false
	}
	if t.ItemType == nil {
		return true
	}
	return t.ItemType.equal(*other.ItemType)
}

func (t TypeInfo) sortKey() string {
	if t.ItemType != nil {
		return string(t.Kind) + ":" + t.ItemType.sortKey()
	}
	return string(t.Kind) + ":" + t.TypeName
}

func Unique(typeInfos []TypeInfo) []TypeInfo {
	var result []TypeInfo
	for _, typeInfo := range typeInfos {
		switch {
		case typeInfo.Kind.IsPrimitive():
		case slices.ContainsFunc(result, typeInfo.equal):
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
		case typeInfo.Kind.IsPrimitive():
		default:
			result = append(result, typeInfo)
		}
	}
	return result
}

func Cleanup(typeInfos []TypeInfo) []TypeInfo {
	return slices.DeleteFunc(typeInfos, func(t TypeInfo) bool {
		return t.Kind == ""
	})
}

func Sort(typeInfos []TypeInfo) {
	slices.SortFunc(typeInfos, func(a, b TypeInfo) int {
		return cmp.Compare(a.sortKey(), b.sortKey())
	})
}

func getUniqueExternalTypeInfos(typeInfos []TypeInfo) []TypeInfo {
	var result []TypeInfo

	for _, typeInfo := range typeInfos {
		switch {
		case typeInfo.Kind.IsPrimitive():
		case slices.ContainsFunc(result, typeInfo.equal):
		default:
			result = append(result, typeInfo)
		}
	}

	result = slices.DeleteFunc(result, func(t TypeInfo) bool {
		return t.Kind == ""
	})

	slices.SortFunc(result, func(a, b TypeInfo) int {
		return cmp.Compare(a.sortKey(), b.sortKey())
	})

	return result
}

func getTypeInfo(ctx context.TemplateGenerationContext, prop *openapi3.SchemaRef) TypeInfo {
	if prop.Value.Type.Includes("array") {
		itemTypeInfo := getTypeInfo(ctx, prop.Value.Items)
		return TypeInfo{
			Kind:     KindArray,
			Layer:    itemTypeInfo.Layer,
			ItemType: &itemTypeInfo,
		}
	}

	// $ref handling:
	// "#/components/schemas/User" -> "User"
	if prop.Ref != "" {
		schemaName := filepath.Base(prop.Ref)
		return TypeInfo{
			Kind:     KindRef,
			TypeName: schemaName,
			Layer:    ctx.Layer,
		}
	}

	switch {
	case prop.Value.Type.IsSingle():
		switch prop.Value.Type.Slice()[0] {
		case "string":
			return TypeInfo{Kind: KindString}
		case "integer":
			return TypeInfo{Kind: KindInteger}
		case "number":
			return TypeInfo{Kind: KindNumber}
		case "boolean":
			return TypeInfo{Kind: KindBoolean}
		case "object":
			if prop.Value.AdditionalProperties.Schema != nil {
				itemTypeInfo := getTypeInfo(ctx, prop.Value.AdditionalProperties.Schema)
				return TypeInfo{
					Kind:     KindMap,
					Layer:    itemTypeInfo.Layer,
					ItemType: &itemTypeInfo,
				}
			}

			return TypeInfo{Kind: KindObject}
		}
	case prop.Value.Format == "date-time":
		return TypeInfo{Kind: KindDateTime}
	}

	return TypeInfo{Kind: KindString}
}
