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

package configuration

import (
	"encoding/json"
	"errors"
	"os"
)

type CoreServiceConfiguration struct {
	Environment string `json:"environment"`
}

type MongoConfiguration struct {
	ClientUri string `json:"clientUri"`
	Database  string `json:"database"`
}

type RedisConfiguration struct {
	Addr string `json:"addr"`
}

type ServiceConfiguration struct {
	CoreServiceConfiguration CoreServiceConfiguration `json:"core"`
	MongoConfiguration       MongoConfiguration       `json:"mongo"`
	RedisConfiguration       RedisConfiguration       `json:"redis"`
}

func LoadServiceConfiguration() (*ServiceConfiguration, error) {
	configPath := os.Getenv("JIJ_SERVICE_CONFIG_URI")
	if configPath == "" {
		return nil, errors.New("JIJ_SERVICE_CONFIG_URI not defined")
	}

	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.New("Failed to read file:" + err.Error())
	}

	var serviceConfiguration ServiceConfiguration
	err = json.Unmarshal(fileBytes, &serviceConfiguration)
	if err != nil {
		return nil, errors.New("Failed to unmarshal JSON:" + err.Error())
	}
	return &serviceConfiguration, nil
}
