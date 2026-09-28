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

package tests

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/analytics"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/devices"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/tests/mocks"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
)

// TestGroupSelectionSendsAspectListsToDeviceRepository reuses the group-with-service-criteria case with criteria
// naming aspects. It asserts the criteria that reach the device-repository rather than the deployed pipeline,
// because the mock answers independently of the criteria and the pipeline would look the same for a wrong one.
func TestGroupSelectionSendsAspectListsToDeviceRepository(t *testing.T) {
	const caseDir = RESOURCE_BASE_DIR + "group-with-service-criteria/"
	const inputId = "373808f2-848a-4446-8062-abd973dc96d3"

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, _, camunda, _, devicerepo, flowparser, _, err := prepareMocks(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}

	var flowModelCells []analytics.FlowModelCell
	readJsonFile(t, caseDir+"flow_model_cells.json", &flowModelCells)
	flowparser.SetResponse(flowModelCells)

	var selectables, selectables2 []devices.DeviceTypeSelectable
	readJsonFile(t, caseDir+"device_type_selectables.json", &selectables)
	readJsonFile(t, caseDir+"device_type_selectables_2.json", &selectables2)
	devicerepo.SetDeviceTypeSelectablesResponse(selectables)
	devicerepo.SetSecondResponse(selectables2)

	var permissionsQueryResponses map[string][]map[string]interface{}
	readJsonFile(t, caseDir+"permissions_query_responses.json", &permissionsQueryResponses)
	devicerepo.SetLegacyPermissionsResponses(permissionsQueryResponses)

	var expectedCamundaRequests []mocks.Request
	readJsonFile(t, caseDir+"expected_camunda_requests.json", &expectedCamundaRequests)

	var tasks []model.CamundaExternalTask
	readJsonFile(t, caseDir+"camunda_tasks.json", &tasks)
	tasks[0].Variables["analytics.criteria."+inputId+".port-name"] = model.CamundaVariable{
		Value: `[{"function_id":"foo","aspect_ids":["a1","a2"]}]`,
	}
	tasks[0].Variables["analytics.service_criteria."+inputId+".port-name"] = model.CamundaVariable{
		Value: `[{"function_id":"foo","aspect_id":"a3"},{"function_id":"bar","aspect_id":"a5","aspect_ids":["a4"]}]`,
	}
	camunda.AddToQueue(tasks)

	time.Sleep(1 * time.Second)

	actualCamundaRequests := camunda.PopRequestLog()
	if !reflect.DeepEqual(expectedCamundaRequests, actualCamundaRequests) {
		e, _ := json.Marshal(expectedCamundaRequests)
		a, _ := json.Marshal(actualCamundaRequests)
		t.Error("task was not completed as without aspects\n", string(e), "\n", string(a))
	}

	actualCriteria := [][]devices.FilterCriteria{}
	for _, request := range devicerepo.PopRequestLog() {
		if request.Endpoint != "/v2/query/device-type-selectables" {
			continue
		}
		var criteria []devices.FilterCriteria
		err = json.Unmarshal([]byte(request.Message), &criteria)
		if err != nil {
			t.Fatal(err)
		}
		actualCriteria = append(actualCriteria, criteria)
	}
	expectedCriteria := [][]devices.FilterCriteria{
		{
			{Interaction: devices.EVENT, FunctionId: "foo", AspectIds: []string{"a1", "a2"}},
		},
		{
			{Interaction: devices.EVENT, FunctionId: "foo", AspectId: "a3", AspectIds: []string{"a3"}},
			{Interaction: devices.EVENT, FunctionId: "bar", AspectId: "a5", AspectIds: []string{"a4", "a5"}},
		},
	}
	if !reflect.DeepEqual(actualCriteria, expectedCriteria) {
		t.Errorf("expected %#v, got %#v", expectedCriteria, actualCriteria)
	}
}

func readJsonFile(t *testing.T, path string, result interface{}) {
	t.Helper()
	file, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(file, result)
	if err != nil {
		t.Fatal(err)
	}
}
