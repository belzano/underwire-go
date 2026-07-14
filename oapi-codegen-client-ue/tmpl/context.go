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
	"os"
	"path/filepath"
	"text/template"
)

type Context struct {
	Layer string
}

func GenerateServiceClientHelpers(ctx context.TemplateGenerationContext) {
	data := Context{
		Layer: ctx.Layer,
	}
	gen := map[string]string{
		ctx.TemplateDir + "/serviceclienthelper.h.tmpl":   "ServiceClientHelper.h",
		ctx.TemplateDir + "/serviceclienthelper.cpp.tmpl": "ServiceClientHelper.cpp",
		ctx.TemplateDir + "/serviceclientmodels.h.tmpl":   "ServiceClientModels.h",
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
