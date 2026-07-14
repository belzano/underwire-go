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

import (
	"log"
	"oapi-codegen-client-ue/context"
	"oapi-codegen-client-ue/oapi"
	"os"
	"path/filepath"
	"text/template"
)

type ServiceClient struct {
	Layer             string
	ServiceClientName string
	Endpoints         []ServiceEndpoint
	ExternalTypes     []TypeInfo
}

func GenerateServiceClient(ctx context.TemplateGenerationContext, endpoints []oapi.ServiceEndpoint) {

	var externalTypes []oapi.TypeInfo
	for _, endpoint := range endpoints {
		for _, depType := range endpoint.GetDependantTypes() {
			externalTypes = append(externalTypes, depType)
		}
	}

	dependencies := oapi.RemoveEngineTypes(externalTypes)
	dependencies = oapi.Unique(dependencies)
	dependencies = oapi.Cleanup(dependencies)
	oapi.Sort(dependencies)

	data := ServiceClient{
		Layer:             ctx.Layer,
		ServiceClientName: "ServiceClient",
		Endpoints:         newServiceEndpoints(endpoints),
		ExternalTypes:     newTypeInfos(dependencies),
	}

	gen := map[string]string{
		ctx.TemplateDir + "/serviceclient.h.tmpl":   "ServiceClient.h",
		ctx.TemplateDir + "/serviceclient.cpp.tmpl": "ServiceClient.cpp",
	}
	for tmplFilePath, outFilePath := range gen {
		tmpl, err := template.ParseFiles(tmplFilePath)
		if err != nil {
			log.Fatalf("failed to load template: %v", err)
		}

		outputFile := filepath.Join(ctx.OutputDir+"/"+ctx.Layer, outFilePath)
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
