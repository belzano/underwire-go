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

package tmpl

import "oapi-codegen-client-ue/oapi"

type TypeInfo struct {
	TypeName              string
	UnrealType            string
	IsEngineType          bool
	Layer                 string
	IsArray               bool
	IsTemplate            bool
	TemplateParamTypeName string
}

func newTypeInfo(t oapi.TypeInfo) TypeInfo {
	unrealType, isTemplate, templateParamTypeName := unrealTypeOf(t)
	return TypeInfo{
		TypeName:              t.TypeName,
		UnrealType:            unrealType,
		IsEngineType:          t.Kind.IsPrimitive(),
		Layer:                 t.Layer,
		IsArray:               t.Kind.IsArray(),
		IsTemplate:            isTemplate,
		TemplateParamTypeName: templateParamTypeName,
	}
}

func newTypeInfos(types []oapi.TypeInfo) []TypeInfo {
	result := make([]TypeInfo, len(types))
	for i, t := range types {
		result[i] = newTypeInfo(t)
	}
	return result
}

// unrealTypeOf derives the Unreal C++ spelling of a framework-agnostic
// oapi.TypeInfo. isTemplate/templateParamTypeName capture that a
// TArray<T>/TMap<FString, T> needs #include "<Layer>/<T>.h" for the inner T,
// not for the (uninclude-able) templated container itself.
func unrealTypeOf(t oapi.TypeInfo) (unrealType string, isTemplate bool, templateParamTypeName string) {
	switch t.Kind {
	case oapi.KindString:
		return "FString", false, ""
	case oapi.KindInteger:
		return "int32", false, ""
	case oapi.KindNumber:
		return "float", false, ""
	case oapi.KindBoolean:
		return "bool", false, ""
	case oapi.KindDateTime:
		return "FDateTime", false, ""
	case oapi.KindObject:
		return "FJsonObjectWrapper", false, ""
	case oapi.KindRef:
		return "F" + t.TypeName, false, ""
	case oapi.KindArray:
		itemUnrealType, _, _ := unrealTypeOf(*t.ItemType)
		return "TArray<" + itemUnrealType + ">", true, t.ItemType.TypeName
	case oapi.KindMap:
		itemUnrealType, _, _ := unrealTypeOf(*t.ItemType)
		return "TMap<FString, " + itemUnrealType + ">", true, t.ItemType.TypeName
	default:
		// The zero-value oapi.TypeInfo{} (Kind == "") is used as a "no
		// request/response body" sentinel by ServiceEndpoint.QueryBodyType/
		// ResponseBodyType — keep UnrealType empty so
		// {{if .QueryBodyType.UnrealType}} in the service client templates
		// still correctly omits the parameter.
		return "", false, ""
	}
}
