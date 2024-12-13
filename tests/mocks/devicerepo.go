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

type DeviceRepo struct {
	requestsLog     []Request
	mux             sync.Mutex
	Response        []devices.DeviceTypeSelectable
	SecondResponse  []devices.DeviceTypeSelectable
	legacyResponses map[string]interface{}
	called          bool
}

func (this *DeviceRepo) SetDeviceTypeSelectablesResponse(value []devices.DeviceTypeSelectable) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.Response = value
}

func (this *DeviceRepo) SetSecondResponse(value []devices.DeviceTypeSelectable) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.SecondResponse = value
}

func (this *DeviceRepo) PopRequestLog() []Request {
	this.mux.Lock()
	defer this.mux.Unlock()
	result := this.requestsLog
	this.requestsLog = []Request{}
	return result
}

func (this *DeviceRepo) logRequest(request *http.Request) {
	this.mux.Lock()
	defer this.mux.Unlock()
	temp, _ := io.ReadAll(request.Body)
	this.requestsLog = append(this.requestsLog, Request{
		Method:   request.Method,
		Endpoint: request.URL.Path,
		Message:  string(temp),
	})
}

func (this *DeviceRepo) logRequestWithMessage(request *http.Request, m interface{}) {
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

func (this *DeviceRepo) Start(ctx context.Context, wg *sync.WaitGroup) (url string) {
	if this.legacyResponses == nil {
		this.legacyResponses = map[string]interface{}{}
	}
	server := httptest.NewServer(this.getRouter())
	wg.Add(1)
	go func() {
		<-ctx.Done()
		server.Close()
		wg.Done()
	}()
	return server.URL
}

func (this *DeviceRepo) getRouter() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		this.logRequest(request)
		if request.Method == "POST" && request.URL.Path == "/v2/query/device-type-selectables" {
			if this.called && this.SecondResponse != nil {
				json.NewEncoder(writer).Encode(this.SecondResponse)
			} else {
				json.NewEncoder(writer).Encode(this.Response)
			}
			this.called = true
			return
		}
		if resp, ok := this.legacyResponses[request.URL.Path]; ok {
			json.NewEncoder(writer).Encode(resp)
			return
		}
		http.Error(writer, "unknown path "+request.URL.Path, 500)
	})
}

func (this *DeviceRepo) SetLegacyPermissionsResponses(responses map[string][]map[string]interface{}) {
	this.legacyResponses = map[string]interface{}{}
	for k, v := range responses {
		this.legacyResponses["/"+k] = v
		for _, e := range v {
			id, ok := e["id"].(string)
			if ok {
				this.legacyResponses["/"+k+"/"+id] = e
			}
		}
	}
}
