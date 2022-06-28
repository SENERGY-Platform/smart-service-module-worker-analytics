/*
 * Copyright (c) 2022 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mocks

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/imports"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

type Import struct {
	requestsLog []Request
	mux         sync.Mutex
}

func (this *Import) PopRequestLog() []Request {
	this.mux.Lock()
	defer this.mux.Unlock()
	result := this.requestsLog
	this.requestsLog = []Request{}
	return result
}

func (this *Import) logRequest(request *http.Request) {
	this.mux.Lock()
	defer this.mux.Unlock()
	temp, _ := io.ReadAll(request.Body)
	this.requestsLog = append(this.requestsLog, Request{
		Method:   request.Method,
		Endpoint: request.URL.Path,
		Message:  string(temp),
	})
}

func (this *Import) logRequestWithMessage(request *http.Request, m interface{}) {
	this.mux.Lock()
	defer this.mux.Unlock()
	message, ok := m.(string)
	if !ok {
		temp, _ := json.Marshal(m)
		message = string(temp)
	}
	this.requestsLog = append(this.requestsLog, Request{
		Method:   request.Method,
		Endpoint: request.URL.Path,
		Message:  message,
	})
}

func (this *Import) Start(ctx context.Context, wg *sync.WaitGroup) (url string) {
	server := httptest.NewServer(this.getRouter())
	wg.Add(1)
	go func() {
		<-ctx.Done()
		server.Close()
		wg.Done()
	}()
	return server.URL
}

func (this *Import) getRouter() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		this.logRequest(request)
		if request.Method == "GET" && strings.HasPrefix(request.URL.Path, "/instances/") {
			json.NewEncoder(writer).Encode(imports.Import{
				KafkaTopic: strings.TrimPrefix(request.URL.Path, "/instances/") + "_topic",
			})
			return
		}
		http.Error(writer, "unknown path", 500)
	})
}
