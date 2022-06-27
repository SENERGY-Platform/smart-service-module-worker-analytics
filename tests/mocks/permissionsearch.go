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
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/devices"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

type PermissionsSearch struct {
	requestsLog []Request
	mux         sync.Mutex
	Response    map[string]interface{}
}

func (this *PermissionsSearch) SetResponse(value map[string]interface{}) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.Response = value
}

func (this *PermissionsSearch) PopRequestLog() []Request {
	this.mux.Lock()
	defer this.mux.Unlock()
	result := this.requestsLog
	this.requestsLog = []Request{}
	return result
}

func (this *PermissionsSearch) logRequest(request *http.Request) {
	this.mux.Lock()
	defer this.mux.Unlock()
	temp, _ := io.ReadAll(request.Body)
	this.requestsLog = append(this.requestsLog, Request{
		Method:   request.Method,
		Endpoint: request.URL.Path,
		Message:  string(temp),
	})
}

func (this *PermissionsSearch) logRequestWithMessage(request *http.Request, m interface{}) {
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

func (this *PermissionsSearch) Start(ctx context.Context, wg *sync.WaitGroup) (url string) {
	server := httptest.NewServer(this.getRouter())
	wg.Add(1)
	go func() {
		<-ctx.Done()
		server.Close()
		wg.Done()
	}()
	return server.URL
}

func (this *PermissionsSearch) getRouter() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		msg, _ := io.ReadAll(request.Body)
		this.logRequestWithMessage(request, string(msg))
		if request.Method == "POST" && request.URL.Path == "/v3/query" {
			query := devices.QueryMessage{}
			json.Unmarshal(msg, &query)
			json.NewEncoder(writer).Encode(this.Response[query.Resource])
			return
		}
		http.Error(writer, "unknown path", 500)
	})
}
