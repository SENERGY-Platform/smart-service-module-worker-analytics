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

package analytics

import (
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/devices"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/configuration"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"sort"
	"strings"
)

func New(config Config, libConfig configuration.Config, auth *auth.Auth, smartServiceRepo SmartServiceRepo, imports Imports, devices Devices) *Analytics {
	return &Analytics{config: config, libConfig: libConfig, auth: auth, smartServiceRepo: smartServiceRepo, imports: imports, devices: devices}
}

type Analytics struct {
	config           Config
	libConfig        configuration.Config
	auth             *auth.Auth
	smartServiceRepo SmartServiceRepo
	imports          Imports
	devices          Devices
}

type Imports interface {
	GetTopic(token auth.Token, importId string) (topic string, err error)
}

type SmartServiceRepo interface {
	GetInstanceUser(instanceId string) (userId string, err error)
	ListExistingModules(processInstanceId string, query model.ModulQuery) (result []model.SmartServiceModule, err error)
}

type Devices interface {
	GetDeviceInfosOfGroup(token auth.Token, groupId string) (devices []devices.Device, deviceTypeIds []string, err error)
	GetDeviceInfosOfDevices(token auth.Token, deviceIds []string) (devices []devices.Device, deviceTypeIds []string, err error)
	GetDeviceTypeSelectables(token auth.Token, criteria []devices.FilterCriteria, includeModified bool, servicesMustMatchAllCriteria bool) (result []devices.DeviceTypeSelectable, err error)
}

func (this *Analytics) Do(task model.CamundaExternalTask) (modules []model.Module, outputs map[string]interface{}, err error) {
	userId, err := this.smartServiceRepo.GetInstanceUser(task.ProcessInstanceId)
	if err != nil {
		log.Println("ERROR: unable to get instance user", err)
		return modules, outputs, err
	}
	token, err := this.auth.ExchangeUserToken(userId)
	if err != nil {
		log.Println("ERROR: unable to exchange user token", err)
		return modules, outputs, err
	}

	outputs = map[string]interface{}{}

	key := this.getModuleKey(task)

	module, returnData, err := this.handleAnalyticsCommand(token, task, key)
	if err != nil {
		return modules, returnData, err
	}

	moduleData := this.getModuleData(task)
	for key, value := range module.ModuleData {
		moduleData[key] = value
	}
	module.ModuleData = moduleData

	modules = append(modules, module)
	for k, v := range returnData {
		outputs[k] = v
	}

	return modules, outputs, err
}

func (this *Analytics) Undo(modules []model.Module, reason error) {
	log.Println("UNDO:", reason)
	for _, module := range modules {
		if module.DeleteInfo != nil {
			err := this.useModuleDeleteInfo(*module.DeleteInfo)
			if err != nil {
				log.Println("ERROR:", err)
				debug.PrintStack()
			}
		}
	}
}

func (this *Analytics) useModuleDeleteInfo(info model.ModuleDeleteInfo) error {
	req, err := http.NewRequest("DELETE", info.Url, nil)
	if err != nil {
		return err
	}
	if info.UserId != "" {
		token, err := this.auth.ExchangeUserToken(info.UserId)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", token.Jwt())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		temp, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	_, _ = io.ReadAll(resp.Body)
	return nil
}

func (this *Analytics) getModuleId(task model.CamundaExternalTask) string {
	return task.ProcessInstanceId + "." + task.Id
}

func (this *Analytics) handleAnalyticsCommand(token auth.Token, task model.CamundaExternalTask, key *string) (module model.Module, outputs map[string]interface{}, err error) {
	if key != nil {
		return this.handleAnalyticsCommandWithKey(token, task, *key)
	} else {
		return this.handleAnalyticsCreate(token, task, []string{})
	}
}

func (this *Analytics) handleAnalyticsCommandWithKey(token auth.Token, task model.CamundaExternalTask, key string) (module model.Module, outputs map[string]interface{}, err error) {
	module, exists, err := this.getExistingModule(task.ProcessInstanceId, key, this.libConfig.CamundaWorkerTopic)
	if !exists {
		return this.handleAnalyticsCreate(token, task, []string{key})
	}
	setModuleUpdateVersion(&module)

	pipelineIdInterface, ok := module.ModuleData["pipeline_id"]
	if !ok {
		log.Printf("WARNING: pipeline-id output not found in module: \n %#v", module)
		return this.handleAnalyticsCreate(token, task, []string{key})
	}
	pipelineId, ok := pipelineIdInterface.(string)
	if !ok {
		err = fmt.Errorf("module device-group-id output is not string: \n %#v", module)
		log.Println("ERROR: ", err)
		return module, outputs, err
	}
	outputs = map[string]interface{}{
		"pipeline_id": pipelineId,
	}

	pipelineRequest, err := this.getPipelineRequest(token, task)
	if err != nil {
		return module, outputs, err
	}
	pipelineRequest.Id = pipelineId

	_, err, _ = this.SendUpdateRequest(token, pipelineRequest)
	if err != nil {
		return module, outputs, err
	}
	return module, outputs, nil
}

func (this *Analytics) handleAnalyticsCreate(token auth.Token, task model.CamundaExternalTask, keys []string) (module model.Module, outputs map[string]interface{}, err error) {
	pipelineRequest, err := this.getPipelineRequest(token, task)
	if err != nil {
		return module, outputs, err
	}

	pipeline, err, _ := this.SendDeployRequest(token, pipelineRequest)
	if err != nil {
		return module, outputs, err
	}

	return model.Module{
			Id:               this.getModuleId(task),
			ProcesInstanceId: task.ProcessInstanceId,
			SmartServiceModuleInit: model.SmartServiceModuleInit{
				DeleteInfo: &model.ModuleDeleteInfo{
					Url:    this.config.FlowEngineUrl + "/pipeline/" + url.PathEscape(pipeline.Id.String()),
					UserId: token.GetUserId(),
				},
				ModuleType: this.libConfig.CamundaWorkerTopic,
				ModuleData: map[string]interface{}{
					"pipeline_id": pipeline.Id.String(),
					"pipeline":    pipeline,
				},
				Keys: keys,
			},
		}, map[string]interface{}{
			"pipeline_id": pipeline.Id.String(),
		}, nil

}

func (this *Analytics) getPipelineRequest(token auth.Token, task model.CamundaExternalTask) (pipelineRequest PipelineRequest, err error) {
	flowId := this.getFlowId(task)
	if flowId == "" {
		err = errors.New("missing flow id")
		return pipelineRequest, err
	}
	inputs, err, _ := this.GetFlowInputs(token, flowId)
	if err != nil {
		return pipelineRequest, err
	}

	pipelineRequest = PipelineRequest{
		FlowId: flowId,
	}

	pipelineRequest.Name = this.getPipelineName(task)
	if pipelineRequest.Name == "" {
		err = errors.New("missing pipeline name")
		return pipelineRequest, err
	}

	pipelineRequest.WindowTime, err = this.getPipelineWindowTime(task)
	if err != nil {
		return pipelineRequest, err
	}

	pipelineRequest.ConsumeAllMessages = this.getConsumeAllMessages(task)

	pipelineRequest.MergeStrategy, err = this.getPipelineMergeStrategy(task)
	if err != nil {
		return pipelineRequest, err
	}

	pipelineRequest.Description = this.getPipelineDescription(task)

	pipelineRequest.Nodes, err = this.inputsToNodes(token, task, inputs)
	if err != nil {
		return pipelineRequest, err
	}

	return pipelineRequest, nil
}

func (this *Analytics) inputsToNodes(token auth.Token, task model.CamundaExternalTask, inputs []FlowModelCell) (result []PipelineNode, err error) {
	for _, input := range inputs {
		node := PipelineNode{
			NodeId:      input.Id,
			Inputs:      nil,
			Config:      nil,
			PersistData: this.getPersistData(task, input.Id),
		}
		for _, conf := range input.Config {
			nodeConf := NodeConfig{
				Name: conf.Name,
			}
			nodeConf.Value, err = this.getPipelineNodeConfig(task, input.Id, conf.Name)
			if err != nil {
				return result, err
			}
			node.Config = append(node.Config, nodeConf)
		}
		for _, port := range input.InPorts {
			selection, err := this.getSelection(task, input.Id, port)
			if err != nil {
				return result, err
			}
			if selection.DeviceSelection == nil && selection.ImportSelection == nil && selection.DeviceGroupSelection == nil {
				continue
			}
			nodeInput, err := this.selectionToNodeInputs(token, selection, task, input.Id, port)
			if err != nil {
				return result, err
			}
			node.Inputs = append(node.Inputs, nodeInput...)
		}

		//group inputs by topic, and filter
		node.Inputs = groupInputs(node.Inputs)

		sort.Slice(node.Inputs, func(i, j int) bool {
			return node.Inputs[i].TopicName < node.Inputs[j].TopicName
		})
		result = append(result, node)
	}
	return result, nil
}

func groupInputs(in []NodeInput) (out []NodeInput) {
	group := map[string][]NodeInput{}
	for _, element := range in {
		key := element.TopicName + "_" + element.FilterType + "_" + element.FilterIds
		group[key] = append(group[key], element)
	}
	out = []NodeInput{}
	for _, element := range group {
		elementInput := NodeInput{
			FilterIds:  "",
			FilterType: "",
			TopicName:  "",
		}
		for _, sub := range element {
			elementInput.Values = append(elementInput.Values, sub.Values...)
			elementInput.FilterType = sub.FilterType
			elementInput.FilterIds = sub.FilterIds
			elementInput.TopicName = sub.TopicName
		}
		out = append(out, elementInput)
	}
	return out
}

func (this *Analytics) selectionToNodeInputs(token auth.Token, selection model.IotOption, task model.CamundaExternalTask, inputId string, portName string) (result []NodeInput, err error) {
	if selection.DeviceSelection != nil {
		if selection.DeviceSelection.ServiceId == nil {
			return this.deviceWithoutServiceSelectionToNodeInputs(token, *selection.DeviceSelection, task, inputId, portName)
		}
		return this.deviceSelectionToNodeInputs(*selection.DeviceSelection, portName)
	}
	if selection.ImportSelection != nil {
		return this.importSelectionToNodeInputs(token, *selection.ImportSelection, portName)
	}
	if selection.DeviceGroupSelection != nil {
		return this.groupSelectionToNodeInputs(token, *selection.DeviceGroupSelection, task, inputId, portName)
	}
	return result, errors.New("expect selection to contain none nil value")
}

func (this *Analytics) deviceSelectionToNodeInputs(selection model.DeviceSelection, inputPort string) (result []NodeInput, err error) {
	if selection.ServiceId == nil {
		return result, errors.New("expect device selection to contain service info")
	}
	if selection.Path == nil {
		return result, errors.New("expect device selection to contain path info")
	}
	path := this.config.DevicePathPrefix + *selection.Path
	return []NodeInput{{
		FilterIds:  selection.DeviceId,
		FilterType: DeviceFilterType,
		TopicName:  ServiceIdToTopic(*selection.ServiceId),
		Values: []NodeValue{{
			Name: inputPort,
			Path: path,
		}},
	}}, nil
}

func (this *Analytics) deviceWithoutServiceSelectionToNodeInputs(token auth.Token, selection model.DeviceSelection, task model.CamundaExternalTask, inputId string, portName string) (result []NodeInput, err error) {
	criteria, err := this.getNodePathCriteria(task, inputId, portName)
	if err != nil {
		return result, err
	}
	serviceIds, serviceToDevices, serviceToPaths, err := this.getServicesAndPathsForDeviceIdList(token, []string{selection.DeviceId}, criteria)
	if err != nil {
		return result, err
	}

	serviceCriteria, err := this.getNodeServiceCriteria(task, inputId, portName)
	if err != nil {
		return result, err
	}
	if len(serviceCriteria) > 0 {
		filterServiceIds, _, _, err := this.getServicesAndPathsForDeviceIdList(token, []string{selection.DeviceId}, serviceCriteria)
		if err != nil {
			return result, err
		}
		serviceIds, serviceToDevices, serviceToPaths = filterServices(serviceIds, serviceToDevices, serviceToPaths, filterServiceIds)
	}

	result = this.serviceInfosToNodeInputs(serviceIds, serviceToDevices, serviceToPaths, portName)
	return result, nil
}

func (this *Analytics) groupSelectionToNodeInputs(token auth.Token, selection model.DeviceGroupSelection, task model.CamundaExternalTask, inputId string, portName string) (result []NodeInput, err error) {
	criteria, err := this.getNodePathCriteria(task, inputId, portName)
	if err != nil {
		return result, err
	}
	serviceIds, serviceToDevices, serviceToPaths, err := this.getServicesAndPathsForGroupSelection(token, selection, criteria)
	if err != nil {
		return result, err
	}

	serviceCriteria, err := this.getNodeServiceCriteria(task, inputId, portName)
	if err != nil {
		return result, err
	}
	if len(serviceCriteria) > 0 {
		filterServiceIds, _, _, err := this.getServicesAndPathsForGroupSelection(token, selection, serviceCriteria)
		if err != nil {
			return result, err
		}
		serviceIds, serviceToDevices, serviceToPaths = filterServices(serviceIds, serviceToDevices, serviceToPaths, filterServiceIds)
	}

	result = this.serviceInfosToNodeInputs(serviceIds, serviceToDevices, serviceToPaths, portName)
	return result, nil
}

func filterServices(ids []string, toDevices map[string][]string, paths map[string][]string, filter []string) (serviceIds []string, serviceToDevices map[string][]string, serviceToPath map[string][]string) {
	serviceIds = []string{}
	serviceToDevices = map[string][]string{}
	serviceToPath = map[string][]string{}
	index := map[string]bool{}
	for _, id := range filter {
		index[id] = true
	}
	for _, id := range ids {
		if index[id] {
			serviceIds = append(serviceIds, id)
		}
	}
	for id, value := range toDevices {
		if index[id] {
			serviceToDevices[id] = value
		}
	}
	for id, value := range paths {
		if index[id] {
			serviceToPath[id] = value
		}
	}
	return
}

func (this *Analytics) serviceInfosToNodeInputs(serviceIds []string, serviceToDevices map[string][]string, serviceToPaths map[string][]string, inputPort string) (result []NodeInput) {
	for _, serviceId := range serviceIds {
		deviceIds := strings.Join(serviceToDevices[serviceId], ",")
		if deviceIds == "" {
			log.Println("WARNING: missing deviceIds for service in serviceInfosToNodeInputs()", serviceId, " --> skip service")
			continue
		}
		paths := serviceToPaths[serviceId]
		if len(paths) == 0 {
			log.Println("WARNING: missing path for service in serviceInfosToNodeInputs()", serviceId, " --> skip service")
			continue
		}
		values := []NodeValue{}
		if this.config.EnableMultiplePaths {
			for _, path := range paths {
				values = append(values, NodeValue{
					Name: inputPort,
					Path: this.config.GroupPathPrefix + path,
				})
			}
		} else {
			values = []NodeValue{{
				Name: inputPort,
				Path: this.config.GroupPathPrefix + paths[0],
			}}
		}

		result = append(result, NodeInput{
			FilterIds:  deviceIds,
			FilterType: DeviceFilterType,
			TopicName:  ServiceIdToTopic(serviceId),
			Values:     values,
		})
	}
	return result
}

func (this *Analytics) importSelectionToNodeInputs(token auth.Token, selection model.ImportSelection, inputPort string) (result []NodeInput, err error) {
	if selection.Id == "" {
		return result, errors.New("expect import selection to contain id")
	}
	if selection.Path == nil {
		return result, errors.New("expect import selection to contain path")
	}
	topic, err := this.imports.GetTopic(token, selection.Id)
	if err != nil {
		return result, fmt.Errorf("unable to get topic for import (%v): %w", selection.Id, err)
	}
	path := *selection.Path
	if this.config.RemoveImportPathRoot {
		_, temp, found := strings.Cut(path, ".")
		if found {
			path = temp
		}
	}
	path = this.config.ImportPathPrefix + path
	return []NodeInput{{
		FilterIds:  selection.Id,
		FilterType: ImportFilterType,
		TopicName:  topic,
		Values: []NodeValue{{
			Name: inputPort,
			Path: path,
		}},
	}}, nil
}

func (this *Analytics) getExistingModule(processInstanceId string, key string, moduleType string) (module model.Module, exists bool, err error) {
	existingModules, err := this.smartServiceRepo.ListExistingModules(processInstanceId, model.ModulQuery{
		KeyFilter:  &key,
		TypeFilter: &moduleType,
	})
	if err != nil {
		log.Println("ERROR:", err)
		return module, false, err
	}
	if this.config.Debug {
		log.Printf("DEBUG: existing module request: %v, %v, %v, \n %#v", processInstanceId, key, moduleType, existingModules)
	}
	if len(existingModules) == 0 {
		return module, false, nil
	}
	if len(existingModules) > 1 {
		log.Printf("WARNING: more than one existing module found: %v, %v, %v, \n %#v", processInstanceId, key, moduleType, existingModules)
	}
	module.SmartServiceModuleInit = existingModules[0].SmartServiceModuleInit
	module.ProcesInstanceId = processInstanceId
	module.Id = existingModules[0].Id
	return module, true, nil
}

const ModuleUpdateVersionField = "module_update_version"

func setModuleUpdateVersion(module *model.Module) {
	version, versionFieldExists := module.ModuleData[ModuleUpdateVersionField]
	if !versionFieldExists {
		module.ModuleData[ModuleUpdateVersionField] = 1
	}
	versionNum, versionIsNum := version.(float64)
	if !versionIsNum {
		module.ModuleData[ModuleUpdateVersionField] = 1
	}
	module.ModuleData[ModuleUpdateVersionField] = versionNum + 1
}

func ServiceIdToTopic(id string) string {
	id = strings.ReplaceAll(id, "#", "_")
	id = strings.ReplaceAll(id, ":", "_")
	return id
}
