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

package tests

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/analytics"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/devices"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/tests/mocks"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/configuration"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

const RESOURCE_BASE_DIR = "./test-cases/"

func TestWithMocks(t *testing.T) {
	infos, err := os.ReadDir(RESOURCE_BASE_DIR)
	if err != nil {
		t.Error(err)
		return
	}
	for _, info := range infos {
		name := info.Name()
		if info.IsDir() && isValidaForMockTest(RESOURCE_BASE_DIR+name) {
			t.Run(name, func(t *testing.T) {
				mockTest(t, name)
			})
		}
	}
}

func prepareMocks(ctx context.Context, wg *sync.WaitGroup) (
	libConf configuration.Config,
	conf analytics.Config,
	camunda *mocks.CamundaMock,
	smartServiceRepo *mocks.SmartServiceRepoMock,
	devicerepo *mocks.DeviceRepo,
	permissions *mocks.PermissionsSearch,
	flowparser *mocks.FlowParser,
	flowengine *mocks.FlowEngine,
	err error,
) {
	libConf, err = configuration.LoadLibConfig("../config.json")
	if err != nil {
		return
	}
	conf, err = configuration.Load[analytics.Config]("../config.json")
	if err != nil {
		return
	}
	libConf.CamundaWorkerWaitDurationInMs = 200

	camunda = mocks.NewCamundaMock()
	libConf.CamundaUrl = camunda.Start(ctx, wg)

	libConf.AuthEndpoint = mocks.Keycloak(ctx, wg)

	devicerepo = &mocks.DeviceRepo{}
	conf.DeviceRepositoryUrl = devicerepo.Start(ctx, wg)

	imports := &mocks.Import{}
	conf.ImportDeployUrl = imports.Start(ctx, wg)

	permissions = &mocks.PermissionsSearch{}
	conf.PermSearchUrl = permissions.Start(ctx, wg)

	flowparser = &mocks.FlowParser{}
	conf.FlowParserUrl = flowparser.Start(ctx, wg)

	flowengine = &mocks.FlowEngine{}
	conf.FlowEngineUrl = flowengine.Start(ctx, wg)

	smartServiceRepo = mocks.NewSmartServiceRepoMock(libConf, conf)
	libConf.SmartServiceRepositoryUrl = smartServiceRepo.Start(ctx, wg)

	err = pkg.Start(ctx, wg, conf, libConf)

	return
}

func isValidaForMockTest(dir string) bool {
	return checkFileExistence(dir, []string{
		"camunda_tasks.json",
		"expected_camunda_requests.json",
		"expected_smart_service_repo_requests.json",
		"device_type_selectables.json",
		"permissions_query_responses.json",
		"flow_model_cells.json",
		"expected_engine_requests.json",
	})
}

func checkFileExistence(dir string, expectedFiles []string) bool {
	infos, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	files := map[string]bool{}
	for _, info := range infos {
		if !info.IsDir() {
			files[info.Name()] = true
		}
	}
	for _, expected := range expectedFiles {
		if !files[expected] {
			return false
		}
	}
	return true
}

func mockTest(
	t *testing.T,
	name string,
) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, _, camunda, repo, devicerepo, permissions, flowparser, flowengine, err := prepareMocks(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	flowModelCellsFile, err := os.ReadFile(RESOURCE_BASE_DIR + name + "/flow_model_cells.json")
	if err != nil {
		t.Error(err)
		return
	}
	var flowModelCells []analytics.FlowModelCell
	err = json.Unmarshal(flowModelCellsFile, &flowModelCells)
	if err != nil {
		t.Error(err)
		return
	}
	flowparser.SetResponse(flowModelCells)

	moduleListResponse, err := os.ReadFile(RESOURCE_BASE_DIR + name + "/module_list_response.json")
	if err == nil {
		repo.SetListResponse(moduleListResponse)
	}

	deviceTypeSelectablesFile, err := os.ReadFile(RESOURCE_BASE_DIR + name + "/device_type_selectables.json")
	if err != nil {
		t.Error(err)
		return
	}
	var deviceTypeSelectables []devices.DeviceTypeSelectable
	err = json.Unmarshal(deviceTypeSelectablesFile, &deviceTypeSelectables)
	if err != nil {
		t.Error(err)
		return
	}
	devicerepo.SetResponse(deviceTypeSelectables)

	if checkFileExistence(RESOURCE_BASE_DIR+name, []string{"device_type_selectables_2.json"}) {
		deviceTypeSelectablesFile2, err := os.ReadFile(RESOURCE_BASE_DIR + name + "/device_type_selectables_2.json")
		if err != nil {
			t.Error(err)
			return
		}
		var deviceTypeSelectables2 []devices.DeviceTypeSelectable
		err = json.Unmarshal(deviceTypeSelectablesFile2, &deviceTypeSelectables2)
		if err != nil {
			t.Error(err)
			return
		}
		devicerepo.SetSecondResponse(deviceTypeSelectables2)
	}

	permissionsQueryResponsesFile, err := os.ReadFile(RESOURCE_BASE_DIR + name + "/permissions_query_responses.json")
	if err != nil {
		t.Error(err)
		return
	}
	var permissionsQueryResponses map[string]interface{}
	err = json.Unmarshal(permissionsQueryResponsesFile, &permissionsQueryResponses)
	if err != nil {
		t.Error(err)
		return
	}
	permissions.SetResponse(permissionsQueryResponses)

	expectedCamundaRequestsFile, err := os.ReadFile(RESOURCE_BASE_DIR + name + "/expected_camunda_requests.json")
	if err != nil {
		t.Error(err)
		return
	}
	var expectedCamundaRequests []mocks.Request
	err = json.Unmarshal(expectedCamundaRequestsFile, &expectedCamundaRequests)
	if err != nil {
		t.Error(err)
		return
	}

	expectedSmartServiceRepoRequestsFile, err := os.ReadFile(RESOURCE_BASE_DIR + name + "/expected_smart_service_repo_requests.json")
	if err != nil {
		t.Error(err)
		return
	}
	var expectedSmartServiceRepoRequests []mocks.Request
	err = json.Unmarshal(expectedSmartServiceRepoRequestsFile, &expectedSmartServiceRepoRequests)
	if err != nil {
		t.Error(err)
		return
	}

	expectedEngineRequestsFile, err := os.ReadFile(RESOURCE_BASE_DIR + name + "/expected_engine_requests.json")
	if err != nil {
		t.Error(err)
		return
	}
	var expectedEngineRequests []mocks.Request
	err = json.Unmarshal(expectedEngineRequestsFile, &expectedEngineRequests)
	if err != nil {
		t.Error(err)
		return
	}

	tasksFile, err := os.ReadFile(RESOURCE_BASE_DIR + name + "/camunda_tasks.json")
	if err != nil {
		t.Error(err)
		return
	}
	var tasks []model.CamundaExternalTask
	err = json.Unmarshal(tasksFile, &tasks)
	if err != nil {
		t.Error(err)
		return
	}
	camunda.AddToQueue(tasks)

	time.Sleep(1 * time.Second)

	actualCamundaRequests := camunda.PopRequestLog()
	if !reflect.DeepEqual(expectedCamundaRequests, actualCamundaRequests) {
		e, _ := json.Marshal(expectedCamundaRequests)
		a, _ := json.Marshal(actualCamundaRequests)
		t.Error("\n", string(e), "\n", string(a))
	}

	actualSmartServiceRepoRequests := repo.PopRequestLog()
	if !reflect.DeepEqual(expectedSmartServiceRepoRequests, actualSmartServiceRepoRequests) {
		e, _ := json.Marshal(expectedSmartServiceRepoRequests)
		a, _ := json.Marshal(actualSmartServiceRepoRequests)
		t.Error("\n", string(e), "\n", string(a))
	}

	actualEngineRequests := flowengine.PopRequestLog()
	if !reflect.DeepEqual(expectedEngineRequests, actualEngineRequests) {
		a, _ := json.Marshal(actualEngineRequests)
		e, _ := json.Marshal(expectedEngineRequests)
		t.Error("\n", string(a), "\n", string(e))
	}
}
