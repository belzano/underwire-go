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

// To update the executable, run:
// GOOS=windows GOARCH=amd64 go build -o oapi-codegen-client-ue.exe

package main

import (
	"log"
	"oapi-codegen-client-ue/context"
	"oapi-codegen-client-ue/oapi"
	"oapi-codegen-client-ue/tmpl"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func main() {

	args := os.Args[1:]
	params := make(map[string]string)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
	}

	spec := params["-spec"]
	log.Println("Generating from [" + spec + "] ")

	swagger, err := openapi3.NewLoader().LoadFromFile(spec)
	if err != nil {
		log.Fatalf("Failed to load OpenAPI definition: %v", err)
	}

	ctx := context.TemplateGenerationContext{
		TemplateDir: params["-tmpl"],
		OutputDir:   params["-out"],
		Layer:       "serviceclient",
	}

	log.Println("Generating with templates [" + ctx.TemplateDir + "] to [" + ctx.OutputDir + "], target layer [" + ctx.Layer + "]")

	for _, dir := range [...]string{ctx.OutputDir, ctx.OutputDir + "/" + ctx.Layer} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("failed to create output dir: %v", err)
		}
	}

	log.Println("* Generating components")
	components := oapi.ExtractComponents(ctx, swagger.Components)
	for _, comp := range components {
		tmpl.GenerateStruct(ctx, comp)
	}

	log.Println("* Generating Service Client")
	endpoints := oapi.ExtractServiceEndpoints(ctx, swagger.Paths)
	tmpl.GenerateServiceClient(ctx, endpoints)

	log.Println("* Generating Service Client Helpers")
	tmpl.GenerateServiceClientHelpers(ctx)

	log.Println("Generation complete")
}
