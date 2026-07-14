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

type ServiceEndpoint struct {
	Name             string
	Verb             string
	PathPrintfStyle  string
	QueryParameters  []Field
	QueryBodyType    TypeInfo
	ResponseBodyType TypeInfo
}

func newServiceEndpoint(e oapi.ServiceEndpoint) ServiceEndpoint {
	return ServiceEndpoint{
		Name:             e.Name,
		Verb:             e.Verb,
		PathPrintfStyle:  e.PathPrintfStyle,
		QueryParameters:  newFields(e.QueryParameters),
		QueryBodyType:    newTypeInfo(e.QueryBodyType),
		ResponseBodyType: newTypeInfo(e.ResponseBodyType),
	}
}

func newServiceEndpoints(endpoints []oapi.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(endpoints))
	for i, e := range endpoints {
		result[i] = newServiceEndpoint(e)
	}
	return result
}
