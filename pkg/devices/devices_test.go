/*
 * Copyright (c) 2026 InfAI (CC SES)
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

package devices

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
)

func TestGetDeviceTypeSelectablesSendsAspectIds(t *testing.T) {
	var received []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v2/query/device-type-selectables" {
			http.Error(writer, "unexpected request "+request.Method+" "+request.URL.Path, http.StatusNotFound)
			return
		}
		err := json.NewDecoder(request.Body).Decode(&received)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = writer.Write([]byte("[]"))
	}))
	defer server.Close()

	_, err := New(server.URL).GetDeviceTypeSelectables(context.Background(), auth.Token{Token: "Bearer token"}, []FilterCriteria{
		{Interaction: EVENT, FunctionId: "f", AspectIds: []string{"a", "b"}},
		{Interaction: EVENT, FunctionId: "f", AspectId: "c", AspectIds: []string{"c"}},
	}, true, true)
	if err != nil {
		t.Fatal(err)
	}

	expected := []map[string]interface{}{
		{"interaction": "event", "function_id": "f", "device_class_id": "", "aspect_id": "", "aspect_ids": []interface{}{"a", "b"}},
		{"interaction": "event", "function_id": "f", "device_class_id": "", "aspect_id": "c", "aspect_ids": []interface{}{"c"}},
	}
	if !reflect.DeepEqual(received, expected) {
		t.Errorf("expected %#v, got %#v", expected, received)
	}
}
