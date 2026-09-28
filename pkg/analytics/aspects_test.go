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

package analytics

import (
	"context"
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/devices"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
)

type selectablesRecorder struct {
	criteria [][]devices.FilterCriteria
}

func (this *selectablesRecorder) GetDeviceInfosOfGroup(ctx context.Context, token auth.Token, groupId string) ([]devices.Device, []string, error) {
	return nil, nil, nil
}

func (this *selectablesRecorder) GetDeviceInfosOfDevices(ctx context.Context, token auth.Token, deviceIds []string) ([]devices.Device, []string, error) {
	return nil, nil, nil
}

func (this *selectablesRecorder) GetDeviceTypeSelectables(ctx context.Context, token auth.Token, criteria []devices.FilterCriteria, includeModified bool, servicesMustMatchAllCriteria bool) ([]devices.DeviceTypeSelectable, error) {
	this.criteria = append(this.criteria, append([]devices.FilterCriteria{}, criteria...))
	return nil, nil
}

func TestGetDeviceGroupPathOptionsSendsAspectLists(t *testing.T) {
	cases := []struct {
		name     string
		criteria devices.FilterCriteria
		expected devices.FilterCriteria
	}{
		{
			name:     "keeps a criteria without aspect free of aspects",
			criteria: devices.FilterCriteria{FunctionId: "f"},
			expected: devices.FilterCriteria{Interaction: devices.EVENT, FunctionId: "f", AspectIds: []string{}},
		},
		{
			name:     "sends a deprecated aspect id as single element list and keeps it for older device-repositories",
			criteria: devices.FilterCriteria{FunctionId: "f", AspectId: "a"},
			expected: devices.FilterCriteria{Interaction: devices.EVENT, FunctionId: "f", AspectId: "a", AspectIds: []string{"a"}},
		},
		{
			name:     "sends an aspect list unchanged",
			criteria: devices.FilterCriteria{FunctionId: "f", AspectIds: []string{"a", "b"}},
			expected: devices.FilterCriteria{Interaction: devices.EVENT, FunctionId: "f", AspectIds: []string{"a", "b"}},
		},
		{
			name:     "adds a deprecated aspect id to the aspect list",
			criteria: devices.FilterCriteria{Interaction: devices.REQUEST, FunctionId: "f", AspectId: "c", AspectIds: []string{"a", "b"}},
			expected: devices.FilterCriteria{Interaction: devices.REQUEST, FunctionId: "f", AspectId: "c", AspectIds: []string{"a", "b", "c"}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			recorder := &selectablesRecorder{}
			a := &Analytics{devices: recorder}
			_, err := a.getDeviceGroupPathOptions(context.Background(), auth.Token{}, []devices.FilterCriteria{c.criteria}, []string{"dt1"})
			if err != nil {
				t.Fatal(err)
			}
			expected := [][]devices.FilterCriteria{{c.expected}}
			if !reflect.DeepEqual(recorder.criteria, expected) {
				t.Errorf("expected %#v, got %#v", expected, recorder.criteria)
			}
		})
	}
}

func TestNodeCriteriaVariablesKeepAspectLists(t *testing.T) {
	a := &Analytics{config: Config{WorkerParamPrefix: "analytics."}}
	task := model.CamundaExternalTask{Variables: map[string]model.CamundaVariable{
		"analytics.criteria.input.port":         {Value: `[{"function_id":"f","aspect_ids":["a","b"]}]`},
		"analytics.service_criteria.input.port": {Value: `[{"function_id":"g","aspect_id":"c"},{"function_id":"h","aspect_ids":["d"]}]`},
	}}

	pathCriteria, err := a.getNodePathCriteria(task, "input", "port")
	if err != nil {
		t.Fatal(err)
	}
	expectedPathCriteria := []devices.FilterCriteria{{FunctionId: "f", AspectIds: []string{"a", "b"}}}
	if !reflect.DeepEqual(pathCriteria, expectedPathCriteria) {
		t.Errorf("path criteria: expected %#v, got %#v", expectedPathCriteria, pathCriteria)
	}

	serviceCriteria, err := a.getNodeServiceCriteria(task, "input", "port")
	if err != nil {
		t.Fatal(err)
	}
	expectedServiceCriteria := []devices.FilterCriteria{{FunctionId: "g", AspectId: "c"}, {FunctionId: "h", AspectIds: []string{"d"}}}
	if !reflect.DeepEqual(serviceCriteria, expectedServiceCriteria) {
		t.Errorf("service criteria: expected %#v, got %#v", expectedServiceCriteria, serviceCriteria)
	}
}
