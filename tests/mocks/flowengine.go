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
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/analytics"
	uuid "github.com/satori/go.uuid"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

type FlowEngine struct {
	requestsLog []Request
	mux         sync.Mutex
}

func (this *FlowEngine) PopRequestLog() []Request {
	this.mux.Lock()
	defer this.mux.Unlock()
	result := this.requestsLog
	this.requestsLog = []Request{}
	return result
}

func (this *FlowEngine) logRequest(request *http.Request) {
	this.mux.Lock()
	defer this.mux.Unlock()
	temp, _ := io.ReadAll(request.Body)
	this.requestsLog = append(this.requestsLog, Request{
		Method:   request.Method,
		Endpoint: request.URL.Path,
		Message:  string(temp),
	})
}

func (this *FlowEngine) logRequestWithMessage(request *http.Request, m interface{}) {
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

func (this *FlowEngine) Start(ctx context.Context, wg *sync.WaitGroup) (url string) {
	server := httptest.NewServer(this.getRouter())
	wg.Add(1)
	go func() {
		<-ctx.Done()
		server.Close()
		wg.Done()
	}()
	return server.URL
}

func (this *FlowEngine) getRouter() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		msg, _ := io.ReadAll(request.Body)
		this.logRequestWithMessage(request, string(msg))
		if request.Method == "DELETE" && strings.HasPrefix(request.URL.Path, "/pipeline/") {
			writer.WriteHeader(200)
			return
		}
		if (request.Method == "POST" || request.Method == "PUT") && request.URL.Path == "/pipeline" {
			pipelineRequest := analytics.PipelineRequest{}
			json.Unmarshal(msg, &pipelineRequest)
			pipeline := analytics.Pipeline{
				Name:        pipelineRequest.Name,
				Description: pipelineRequest.Description,
			}
			pipeline.Id, _ = uuid.FromString("1e138d25-d5ee-4a89-9a83-630f4308941a")
			json.NewEncoder(writer).Encode(pipeline)
			return
		}
		http.Error(writer, "unknown path", 500)
	})
}
