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

package main

import (
	"jij-service/api"
	"jij-service/configuration"
	"jij-service/dal"
	"jij-service/domain"
	"jij-service/middleware"
	"log"
	"net/http"
	"os"
)

func main() {

	ServiceConfiguration, err := configuration.LoadServiceConfiguration()
	if err != nil {
		log.Fatalf("failed to load service configuration: %v", err)
	}

	apiServer := new(api.Server{
		ProfileService: domain.ProfileService{
			ServiceConfiguration: *ServiceConfiguration,
			ProfileDal:           dal.NewProfileDalMongo(ServiceConfiguration.MongoConfiguration),
		},
	})

	strictHandler := api.NewStrictHandler(apiServer, nil)

	handler := api.HandlerFromMux(strictHandler, nil)
	handler = middleware.LoggingMiddleware(handler)
	log.Printf("With logging mdw")

	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	server := &http.Server{
		Addr:     addr,
		Handler:  handler,
		ErrorLog: log.New(os.Stderr, "HTTP Server Error: ", log.LstdFlags),
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
